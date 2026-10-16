// A scoped reader. Project metadata (including domain names) is a separate grant.
permit (
    principal in Flyte::Group::"payments-readers",
    action in [Flyte::Action::"RunRead", Flyte::Action::"TaskRead",
               Flyte::Action::"TriggerRead", Flyte::Action::"DataRead",
               Flyte::Action::"LogsRead"],
    resource == Flyte::Scope::"payments/production"
);
permit (
    principal in Flyte::Group::"payments-readers",
    action == Flyte::Action::"ProjectRead",
    resource == Flyte::Project::"payments"
);
permit (
    principal in Flyte::Group::"payments-readers",
    action in [Flyte::Action::"Translate", Flyte::Action::"ClusterRead"],
    resource == Flyte::Deployment::"default"
);

// A scheduler is a principal from the same trusted issuer, with explicit grants.
permit (
    principal == Flyte::Principal::"flyte-scheduler",
    action in [Flyte::Action::"RunCreate", Flyte::Action::"TriggerExecute"],
    resource == Flyte::Scope::"payments/production"
);
permit (
    principal == Flyte::Principal::"flyte-scheduler",
    action == Flyte::Action::"ProjectRead",
    resource == Flyte::Project::"payments"
);

// Runtime reporting has deployment-wide authority. Never grant this to end users.
permit (
    principal == Flyte::Principal::"flyte-runtime",
    action == Flyte::Action::"RuntimeWrite",
    resource == Flyte::Deployment::"default"
);
