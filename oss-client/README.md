# Union.ai Console V2

> Next generation console interface for Union.ai

## 🚀 Quick Start

### Prerequisites

- [Node.js](https://nodejs.org/) (v20 or later) - JavaScript runtime environment required for running the application
- [pnpm](https://pnpm.io/) (v10.8.1) - Fast and efficient package manager alternative to npm
- Go development environment (for backend services)
- Docker (optional, for container testing)

### Installing Node.js *(if you haven't already)*

1. Using Node Version Manager (nvm) - Recommended:
   - Follow the [nvm installation guide](https://github.com/nvm-sh/nvm#installing-and-updating)
   - After installation, install and use Node.js LTS:
     ```bash
     nvm install --lts
     nvm use --lts
     node --version
     ```

2. Direct download:
   - Visit [Node.js Downloads](https://nodejs.org/)
   - Download and install the LTS version (recommended for most users)
   - Verify installation:
     ```bash
     node --version
     # Should output v20.x.x (current LTS)
     ```

### Installing pnpm *(if you haven't already)*

1. Using Corepack (recommended, included with Node.js):
   ```bash
   # Enable Corepack (only needed once after Node.js installation)
   corepack enable

   # Install the required pnpm version
   corepack prepare pnpm@10.8.1 --activate

   # Use the correct pnpm version
   corepack use pnpm@10.8.1
   ```

2. Using npm (alternative method):
   ```bash
   npm install -g pnpm@10.8.1
   ```

3. Verify installation:
   ```bash
   pnpm --version
   # Should output: 10.8.1
   ```

If you see a different version after running `pnpm --version`, make sure to follow the installation steps above to get the correct version.


## 💻 Development Setup

1. Install dependencies:
   ```bash
   pnpm install
   ```

2. Choose your development path:

   #### Option A: Local Development with Devbox
   If you want to develop against a local backend:

   1. Set up the local development environment by following our [setup guide](https://www.notion.so/Running-locally-1c98cc06513d8034b690cca368185e71)



   2. Create a test run:

      * clone [flyte-sdk](https://github.com/flyteorg/flyte-sdk)
      *  Ensure your flyte-sdk `config.yaml` looks like this

      ```bash
      admin:
         endpoint: dns:///localhost:8090
         insecure: true
      image:
       builder: local
      task:
         domain: development
         org: testorg
         project: testproject
       ```
      * create a test run
      ```bash
      flyte run examples/basics/types/int_collection.py main
      ```

   3. Note the output containing the run ID:
      ```
      api_test.go:398: Created root run: org:"testorg" project:"testproject" domain:"development" root_name:"testroot-1744821254" name:"testroot-1744821254"
      ```
      Save these parameters for constructing the client URL.

   4. Start the development server:
      ```bash
      pnpm run dev
      ```
      When prompted, select `localhost`. This will:
      - Start the development server on port 8080
      - Connect to the local backend at http://localhost:8090
      - No hosts file modification needed

   #### Option B: Development Against Remote Domains
   If you want to develop against a specific domain:

   1. Update your `/etc/hosts` file to include the required domain entries:
      ```bash
      sudo vi /etc/hosts
      ```
      Add entries for the domains you need (examples):

         ```bash
         127.0.0.1 localhost.acme.cloud-staging.union.ai
         127.0.0.1 localhost.cloud-staging.union.ai
         127.0.0.1 localhost.demo-gcp.hosted.unionai.cloud
         127.0.0.1 localhost.demo.hosted.unionai.cloud
         127.0.0.1 localhost.dogfood-gcp.cloud-staging.union.ai
         127.0.0.1 localhost.dogfood.cloud-staging.union.ai
         127.0.0.1 localhost.hosted.cloud-staging.union.ai
         127.0.0.1 localhost.playground-gcp.canary.unionai.cloud
         127.0.0.1 localhost.playground.canary.unionai.cloud
         127.0.0.1 localhost.sample-tenant-us-west-2.us-west-2.union.ai
         127.0.0.1 localhost.sample-tenant.canary.unionai.cloud
         127.0.0.1 localhost.sample-tenant.cloud-staging.union.ai
         127.0.0.1 localhost.sample-tenant.hosted.unionai.cloud
         127.0.0.1 localhost.serving-mvp.us-west-2.union.ai
         127.0.0.1 localhost.union-internal.hosted.unionai.cloud
         # Add other domains as needed from the list below
         ```

   2. Start the client:
      ```bash
      pnpm run dev
      ```
      When prompted, select your target domain (e.g., `dogfood.cloud-staging.union.ai`). This will:
      - Generate SSL certificates
      - Start the development server
      - Connect to the selected remote backend


3. Access the application:
   - For localhost: `http://localhost:8080/v2`
   - For specific domains: `https://localhost.{selected-domain}:8080/v2`

   Example URLs:
   ```
   http://localhost:8080/v2/domain/development/project/testproject/runs
   # or
   https://localhost.dogfood.cloud-staging.union.ai:8080/v2/domain/development/project/flytesnacks/runs/
   ```

## 📝 Development

The application source code is located in the `/src` directory. The development server features hot-reload, so changes will be reflected immediately in the browser.

## 📚 Storybook Development

Storybook is our primary tool for component development, testing, and documentation. All stories are located in the `/src/stories` directory.

### Running Storybook

```bash
# Start Storybook development server
pnpm run storybook
# Access at http://localhost:6006

# Build Storybook for production
pnpm run build-storybook

# Serve the production build
pnpm run serve-storybook
```

### Writing New Stories

1. Create a new story file in `/src/stories` following the naming convention `ComponentName.stories.tsx`
2. Use the Component Story Format:
   ```typescript
   import type { Meta, StoryObj } from '@storybook/react'
   import { YourComponent } from '@/components/YourComponent'

   const meta: Meta<typeof YourComponent> = {
     title: 'Components/YourComponent',
     component: YourComponent,
     tags: ['autodocs'],
     parameters: {
       docs: {
         description: {
           component: 'Component description...'
         }
       }
     }
   }

   export default meta
   type Story = StoryObj<typeof YourComponent>

   export const Default: Story = {
     args: {
       // component props
     }
   }
   ```

### Best Practices
- Keep stories focused on one component
- Document all component props and their usage
- Include examples of different states and variations
- Use args to make stories interactive
- Add helpful descriptions and documentation
- Test components in isolation


## 🛠 Tech Stack

- [Next.js](https://nextjs.org/docs) - React framework for production
- [Tailwind CSS](https://tailwindcss.com/docs) - Utility-first CSS framework
- [Headless UI](https://headlessui.dev) - Unstyled UI components
- [Framer Motion](https://www.framer.com/docs/) - Animation library
- [Zustand](https://docs.pmnd.rs/zustand/getting-started/introduction) - State management
- [TanStack Query](https://tanstack.com/query/latest) - Powerful async state management for fetching, caching, and updating data


## 🐳 Docker Testing

### Main Application
```bash
# Build the image
docker build -t union-console --build-arg BUILDKITE_COMMIT=$(git rev-parse HEAD) .

# Run the container
docker run -p 8080:8080 union-console

# Access at http://localhost:8080
```

### Storybook
```bash
# Build Storybook image
docker build -f Dockerfile.storybook -t my-storybook .

# Run Storybook container
docker run -p 8080:8080 my-storybook

# Access at http://localhost:8080
```

## 📄 License

Proprietary and confidential. Unauthorized copying of this repository, via any medium, is strictly prohibited. All rights reserved by Union.ai.

---
Built by Union.ai
