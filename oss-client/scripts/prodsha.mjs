if (!process.env.GIT_SHA) {
    // fail CI/CD if no gitsha is set
    throw new Error('The GIT_SHA environment variable is not set or is falsy.');
} else {
    console.log(`📦 Building CDN with: ${process.env.GIT_SHA}`);
}
