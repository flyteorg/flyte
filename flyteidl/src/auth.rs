pub mod auth {

    use openssl::ssl::{SslConnector, SslMethod, SslStream, SslVerifyMode};
    use std::net::{TcpListener, TcpStream};
    use url::Url;

    use crate::flyteidl::service::{
        auth_metadata_service_client::AuthMetadataServiceClient, OAuth2MetadataRequest,
        PublicClientAuthConfigRequest,
    };
    use anyhow::Result;
    use keyring::Entry;
    use oauth2::{
        basic::BasicClient,
        // Please make sure `ureq` feature flag is enabled. FYR: https://docs.rs/oauth2/latest/oauth2/#importing-oauth2-selecting-an-http-client-interface
        ureq::http_client,
        AccessToken,
        AuthUrl,
        AuthorizationCode,
        ClientId,
        ClientSecret,
        CsrfToken,
        PkceCodeChallenge,
        RedirectUrl,
        RevocationUrl,
        Scope,
        StandardRevocableToken,
        TokenResponse,
        TokenUrl,
    };
    use tokio::runtime::{Builder, Runtime};
    use tonic::{
        transport::{Certificate, Channel, ClientTlsConfig, Endpoint, Uri},
        Request, Response, Status,
    };

    use std::env;
    use std::io::{BufRead, BufReader, Write};
    use std::sync::Arc;

    #[derive(Clone, Default)]
    pub struct Credential {
        access_token: String,
        refresh_token: String,
    }

    impl Credential {
        fn default() -> Credential {
            Credential {
                access_token: "".into(),
                refresh_token: "".into(),
            }
        }
    }

    #[derive(Clone, Default)]
    struct ClientConfig {
        server_endpoint: String,
        insecure: bool,
        token_endpoint: String,
        authorization_endpoint: String,
        redirect_uri: String,
        client_id: String,
        device_authorization_endpoint: String,
        scopes: Vec<String>,
        header_key: String,
        audience: String,
    }

    // TODO: Shorthand struct default because of some members' fields don't have `Default` trait.
    pub struct OAuthClient {
        auth_service: AuthMetadataServiceClient<Channel>,
        runtime: Runtime,
        client_config: ClientConfig,
        // oauth2_metadata: OAuth2Metadata,
        credentials: Credential,
        oauth_client: oauth2::Client<
            oauth2::StandardErrorResponse<oauth2::basic::BasicErrorResponseType>,
            oauth2::StandardTokenResponse<
                oauth2::EmptyExtraTokenFields,
                oauth2::basic::BasicTokenType,
            >,
            oauth2::basic::BasicTokenType,
            oauth2::StandardTokenIntrospectionResponse<
                oauth2::EmptyExtraTokenFields,
                oauth2::basic::BasicTokenType,
            >,
            StandardRevocableToken,
            oauth2::StandardErrorResponse<oauth2::RevocationErrorResponseType>,
        >,
    }

    pub fn bootstrap_uri_from_endpoint(endpoint: &str, insecure: &bool) -> Uri {
        let endpoint_uri: Uri = {
            let protocol: &str = if *insecure { "http" } else { "https" };
            let formatted_uri: String = format!("{}://{}", protocol, endpoint);
            match formatted_uri.parse::<Uri>() {
                Ok(uri) => uri,
                Err(error) => panic!(
                    "Got invalid endpoint when parsing endpoint_uri: {:?}",
                    error
                ),
            }
        };
        endpoint_uri
    }

    // Helper function on retrieve the certificate from the server at the specified address,
    // and return it as a PEM-encoded string.
    pub fn bootstrap_creds_from_server(endpoint_uri: &Uri) -> Certificate {
        // After we bootstraped the first credential from remote server on chain of certificates,
        // We let the whole process of SSL handshakes complete via enable `tls-roots` feature flag or other flags support that.
        // FYR on Handshake Protocol: https://www.rfc-editor.org/rfc/rfc5246#section-7.4.2
        // It works on these features mentioned because they handle the whole SSL handshake,
        // while `tls` flag doesn't, meaning `tls` just handle only the first certificate on chain of certificates.

        let hostname = match endpoint_uri.host() {
            Some(ref host) => format!("{}", host),
            None => panic!("endpoint not found for bootstrapping certificates."),
        };
        // server_address = hostname + port
        let server_address = match endpoint_uri.port() {
            Some(ref port) => format!("{}:{}", &hostname, port),
            // Unable to determine port, setting port to 443
            None => format!("{}:{}", hostname, "443"),
        };

        // Create a connector that doesn't verify the server's certificate
        let mut connector_builder = SslConnector::builder(SslMethod::tls()).unwrap();
        connector_builder.set_verify(SslVerifyMode::NONE);
        // Initialize insecure TCP connection
        let connector = connector_builder.build();

        // Connect to the server and extract the TLS session
        let stream: TcpStream = TcpStream::connect(server_address).unwrap();
        let ssl_stream: SslStream<TcpStream> = connector.connect(&hostname, stream).unwrap();
        // Get the server certificate
        let cert: openssl::x509::X509 = ssl_stream.ssl().peer_certificate().unwrap();
        let pem: Vec<u8> = cert.to_pem().unwrap();
        // return it as a PEM-encoded string.
        let cert_pem: String = String::from_utf8(pem).unwrap();

        // Log it should show text starts with `-----BEGIN CERTIFICATE-----`
        // println!("cert = {cert_pem}");
        // Persists the `.pem`
        // let mut file =
        //     File::create("./ca_cert.pem").unwrap();
        // file.write_all(format!("{cert_pem}").as_bytes()).unwrap();
        // let local_pem =
        // std::fs::read_to_string("./ca_cert.pem")
        //     .unwrap();

        let cert: Certificate = Certificate::from_pem(cert_pem);

        // Retrurn the Tonic certificate
        cert
    }

    // OAuth2 Authorization Client
    impl OAuthClient {
        pub fn new(
            endpoint: &str,
            insecure: &bool,
            client_id: Option<&str>,
            client_credentials_secret: Option<&str>,
        ) -> OAuthClient {
            // Check details for constructing Tokio asynchronous `runtime`: https://docs.rs/tokio/latest/tokio/runtime/struct.Builder.html#method.new_current_thread
            let rt = match Builder::new_current_thread().enable_all().build() {
                Ok(rt) => rt,
                Err(error) => panic!("Failed to initiate Tokio multi-thread runtime: {:?}", error),
            };

            let endpoint = Arc::new(endpoint);

            let endpoint_uri: Uri = bootstrap_uri_from_endpoint(*endpoint, &insecure);

            let cert = bootstrap_creds_from_server(&endpoint_uri);

            let tls = ClientTlsConfig::new()
                .ca_certificate(cert)
                .domain_name(*endpoint.clone());

            let channel = match rt.block_on(
                Channel::builder(endpoint_uri)
                    .tls_config(tls)
                    .unwrap()
                    .connect(),
            ) {
                Ok(ch) => ch,
                Err(error) => panic!(
                    "Failed at connecting to endpoint when constructing channel: {:?}",
                    error
                ),
            };
            let mut auth_metadata_stub: AuthMetadataServiceClient<Channel> =
                AuthMetadataServiceClient::new(channel);

            // Public OAuth Client Config
            let public_client_auth_config_request: PublicClientAuthConfigRequest =
                PublicClientAuthConfigRequest {};
            let req: Request<PublicClientAuthConfigRequest> =
                Request::new(public_client_auth_config_request);
            let public_cfg_res: crate::flyteidl::service::PublicClientAuthConfigResponse =
                (match rt.block_on(auth_metadata_stub.get_public_client_config(req)) {
                    Ok(res) => res,
                    Err(error) => panic!(
                        "Failed at awaiting response from gRPC service server: {:?}",
                        error
                    ),
                })
                .into_inner();

            let oauth2_metadata_request: OAuth2MetadataRequest = OAuth2MetadataRequest {};
            let req = Request::new(oauth2_metadata_request);
            let oauth_mtdata_res = (match rt.block_on(auth_metadata_stub.get_o_auth2_metadata(req))
            {
                Ok(res) => res,
                Err(error) => panic!(
                    "Failed at awaiting response from gRPC service server: {:?}",
                    error
                ),
            })
            .into_inner();
            let client_config: ClientConfig = ClientConfig {
                server_endpoint: (*endpoint).to_string(),
                insecure: *insecure,
                token_endpoint: oauth_mtdata_res.token_endpoint,
                authorization_endpoint: oauth_mtdata_res.authorization_endpoint,
                redirect_uri: public_cfg_res.redirect_uri,
                client_id: public_cfg_res.client_id,
                scopes: public_cfg_res.scopes,
                header_key: public_cfg_res.authorization_metadata_key,
                device_authorization_endpoint: oauth_mtdata_res.device_authorization_endpoint,
                audience: public_cfg_res.audience,
            };

            let mut credentials = Credential::default();
            // Create an OAuth2 client (auth0 from Okta) by specifying the client ID, client secret, authorization URL and token URL.
            let client = BasicClient::new(
                // (client_config.clone().client_id)
                ClientId::new(
                    client_id
                        .map(|s| s.to_string())
                        .unwrap_or_else(|| env::var("CLIENT_ID").unwrap_or_default()),
                ),
                Some(ClientSecret::new(
                    client_credentials_secret
                        .map(|s| s.to_string())
                        .unwrap_or_else(|| env::var("CLIENT_SECRET").unwrap_or_default()),
                )),
                AuthUrl::new((client_config.clone()).authorization_endpoint).unwrap(),
                // Be careful that the `TokenUrl` endpoint in the official documeantion is `<BASE_DOMAIN>/token`.
                // FYR: https://docs.rs/oauth2/latest/oauth2/#example-synchronous-blocking-api
                // In Auth0, it's `<BASE_URL>/oauth/token`.
                // In Okta, it's `<BASE_URL>/v1/token`.
                // It might vary in different IdP, better to use `default` if it's not specify by OAuth2metadata.
                Some(TokenUrl::new((client_config.clone()).token_endpoint).unwrap()),
            )
            // Set the URL the user will be redirected to after the authorization process.
            .set_redirect_uri(RedirectUrl::new((client_config.clone()).redirect_uri).unwrap());
            // We don't need refresh token at present client side context.
            // .set_revocation_uri(
            //     RevocationUrl::new(
            //         todo!()
            //     )
            //     .expect("Invalid revocation endpoint URL"),
            // );

            OAuthClient {
                auth_service: auth_metadata_stub,
                runtime: rt,
                client_config: client_config,
                credentials: credentials.clone(),
                oauth_client: client,
            }
        }

        // Client Credential flow
        pub fn client_secret_authenticate(&mut self) -> Result<String> {
            let pub_cfg: ClientConfig = self.client_config.clone();

            println!(
                "PublicClientAuthConfig.server_endpoint:\t{}\nPublicClientAuthConfig.redirect_uri:\t{}\nPublicClientAuthConfig.client_id:\t{}\nOAuth2Metadata.token_endpoint:\t{}\nOAuth2Metadata.authorization_endpoint:\t{}",
                (pub_cfg.clone()).server_endpoint,
                (pub_cfg.clone()).redirect_uri,
                (pub_cfg.clone()).client_id,
                (pub_cfg.clone()).token_endpoint,
                (pub_cfg.clone()).authorization_endpoint,
            );

            // Client Credential flow
            let token_response = self
                .oauth_client
                .exchange_client_credentials()
                // .add_scope(Scope::new("okta.myAccount.read".to_string()))
                // .add_scope(Scope::new("all".to_string()))
                // .add_scope(Scope::new("offline".to_string()))
                .request(&http_client)
                .unwrap();

            let access_token = token_response.access_token().secret();

            println!("\n\nOAuth2 Authenticator:\nreturned the following\n - access_token:\t{access_token:?}\nthrough\n - AuthMode:\tClientSecret\n\n");

            // Should return `access_token` to be attached for gRPC interceptor by header.
            // Just like what we did in flytekit remote `auth_interceptor.py` L#36 `auth_metadata = self._authenticator.fetch_grpc_call_auth_metadata()`

            let credentials_for_endpoint = &self.client_config.server_endpoint.as_str();
            let credentials_access_token_key = "access_token";
            let entry = Entry::new(credentials_for_endpoint, credentials_access_token_key)?;
            match entry.set_password(access_token) {
                Ok(()) => println!("KeyRing set successfully."),
                Err(err) => println!("KeyRing set not available."),
            };

            Ok(access_token.clone())
        }

        // PKCE flow
        pub fn pkce_authenticate(&mut self) -> Result<String> {
            // Generate a PKCE challenge.
            let (pkce_challenge, pkce_verifier) = PkceCodeChallenge::new_random_sha256();

            // Generate the full authorization URL.
            let (auth_url, csrf_state) = self
                .oauth_client
                .authorize_url(CsrfToken::new_random)
                // Set the desired scopes.
                // .add_scope(Scope::new("offline_access".to_string()))
                // .add_scope(Scope::new("offline".to_string()))
                // .add_scope(Scope::new("all".to_string()))
                // .add_scope(Scope::new("openid".to_string()))
                // Set the PKCE code challenge.
                .set_pkce_challenge(pkce_challenge)
                .url();

            // This is the URL you should redirect the user to, in order to trigger the authorization process.
            println!("Browse to: {}", auth_url);

            // Once the user has been redirected to the redirect URL, they'll have access to the
            // authorization code. For security reasons, your code should verify that the `state`
            // parameter returned by the server matches `csrf_state`.

            // Bootstrap a Oauth http server to listen to redirect_url
            let (code, state) = {
                // A very naive implementation of the redirect server.
                // Prepare the callback server in the background that listen to our flyeadmin endpoint.
                let listener = TcpListener::bind("localhost:53593").unwrap();

                // The server will terminate itself after collecting the first code.
                let Some(mut stream) = listener.incoming().flatten().next() else {
                    panic!("listener terminated without accepting a connection");
                };

                let mut reader = BufReader::new(&stream);

                let mut request_line = String::new();
                reader.read_line(&mut request_line).unwrap();

                let redirect_url = request_line.split_whitespace().nth(1).unwrap(); // TODO: add error handling
                let url = Url::parse(&("http://example.com".to_string() + redirect_url)).unwrap();

                let code = url
                    .query_pairs()
                    .find(|(key, _)| key == "code")
                    .map(|(_, code)| AuthorizationCode::new(code.into_owned()))
                    .unwrap();

                let state = url
                    .query_pairs()
                    .find(|(key, _)| key == "state")
                    .map(|(_, state)| CsrfToken::new(state.into_owned()))
                    .unwrap();

                let message = "Go back to your terminal :)";
                let response = format!(
                    "HTTP/1.1 200 OK\r\ncontent-length: {}\r\n\r\n{}",
                    message.len(),
                    message
                );
                stream.write_all(response.as_bytes()).unwrap();

                (code, state)
            };

            println!(
                "PKCE Authenticator returned the following code:\n{}\n",
                code.secret()
            );
            println!(
                "PKCE Authenticator returned the following state:\n{} (expected `{}`)\n",
                state.secret(),
                csrf_state.secret()
            );

            // Exchange the code with a token.
            // Now you can trade it for an access token.
            let token_response = self
                .oauth_client
                .exchange_code(code)
                // Send the PKCE code verifier in the token request
                .set_pkce_verifier(pkce_verifier)
                .request(&http_client)
                .unwrap();

            // let token_to_revoke: StandardRevocableToken = match token_response.refresh_token() {
            //     Some(token) => token.into(),
            //     None => token_response.access_token().into(), // TODO: mitigate ambiguous token
            // };

            Ok(token_response.access_token().secret().clone())
        }
    }
}
