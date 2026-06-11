# Quickstart Validation Guide

Validate that the AI IDE integration architecture has been applied successfully.

## Scenario 1: Validate Pointer Redirects

1. Ask your current AI IDE to summarize the core project rules or display the project's rules index.
2. The AI should state that it is reading `.agent/rules/00-index.md` (even if you did not explicitly ask it to read that file).

## Scenario 2: Validate Workflow Relocation

1. Check that the `.agents` directory has been removed:
   ```bash
   ls -la .agents
   ```
   **Expected Outcome:** `ls: .agents: No such file or directory`

2. Check that `.agent/workflows` contains the Spec Kit workflows:
   ```bash
   ls -la .agent/workflows | grep speckit
   ```
   **Expected Outcome:** Lists multiple `.md` files like `speckit.plan.md`, `speckit.specify.md`, etc.

## Scenario 3: Execution Validation

1. Run any Spec Kit slash command (e.g., `/speckit.analyze`) in your AI IDE's chat.
2. The AI should instantly recognize the command via `.agent/workflows/speckit.analyze.md` and begin executing the validation process.
