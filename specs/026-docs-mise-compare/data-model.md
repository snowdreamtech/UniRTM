# Data Model: Documentation Upgrade

Since this feature revolves around static documentation in Markdown format powered by VitePress, there are no traditional database schemas or application-level data models to define.

However, the "Content Model" for the new documentation section is as follows:

## Content Entity: `ComparisonSection`

Each tool being compared against UniRTM will follow this structured content model:

- **Target Tool**: The name of the tool (e.g., `mise`, `nvm`).
- **Primary Function**: What the tool is traditionally used for (e.g., "Node.js version management").
- **Architectural Difference**: How UniRTM solves the problem differently (e.g., "Native Go binary vs Bash script hooks").
- **Migration Value**: The direct benefit of migrating to UniRTM (e.g., "Zero shell startup penalty, zero global environment pollution").
- **Usage Equivalence**: A brief table or snippet showing the old command vs the UniRTM approach (e.g., `nvm use 20` vs updating `.unirtm.toml`).
