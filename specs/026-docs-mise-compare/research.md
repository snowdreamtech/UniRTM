# Feature Research: Documentation Upgrade

## Unknowns Resolved

### Deep Dive into `mise` Documentation Structure

- **Decision**: The UniRTM documentation will structurally adopt and surpass the comprehensiveness of `mise`'s documentation, specifically in the areas of architecture, caching, CLI exhaustiveness, and environment management.
- **Rationale**: User feedback (`speckit.clarify`) explicitly pointed out that the previous UniRTM documentation was too vague ("假大空") and superficial compared to `mise`'s deep technical documentation (e.g., 16KB architecture page, exhaustive 60+ file CLI index).
- **Alternatives considered**: Initially, a high-level marketing overview was used. This was rejected because it failed to address the technical depth required by technical evaluators and architects.

### Visual Representation of Architecture

- **Decision**: We will utilize Mermaid.js flowcharts and structural diagrams directly embedded within the markdown files (`architecture.md`, `cache.md`, etc.).
- **Rationale**: Complex concepts like the "100% native Go architecture," "zero shell-pollution," and internal caching algorithms are best explained with visual aids, answering the user's request for "流程图，结构图" (flowcharts, structural diagrams).
- **Alternatives considered**: Static images. Rejected because Mermaid diagrams are easier to maintain, version-control, and render cleanly in VitePress.

### Positioning External Security Integrations

- **Decision**: Security tools like Trivy, Syft, and Gitleaks will be strictly labeled as "External Integrations" (无缝安全集成) rather than native UniRTM features ("天生安全").
- **Rationale**: The user correctly pointed out that claiming these external tools as native features is misleading and exaggerated marketing. Accurate technical representation builds trust.
- **Alternatives considered**: Keeping the "Secure by Default" marketing phrasing. Rejected due to user feedback.
