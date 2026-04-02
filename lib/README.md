# Wizard Engine (lib)

This directory contains the decoupled core logic of Vibe Scaffold. It is designed to be frontend-agnostic, allowing the same AI orchestration to be used by the Web, TUI, and CLI interfaces.

## Core Components

### `AIService.ts`
The primary interface for AI interactions.
- **`streamChat`**: Handles multi-turn conversations with context injection from previous steps.
- **`streamDocument`**: Constructs the final prompts for document generation (One Pager, Dev Spec, etc.).

### `steps/`
Defines the structure of the 4-step wizard.
- `index.ts`: Unified export of all step configurations.
- `stepN-config.ts`: Individual prompts, instructions, and input requirements for each stage.

### `utils/`
- `ai.ts`: Provider initialization (supports custom Base URLs).
- `stepAccess.ts`: Logic to determine if a user can navigate to a specific step.
- `sampleDocs.ts`: Large-scale sample data for testing.

## Integration Example

To use the "Wizard Engine" in a new frontend:

```typescript
import { AIService } from '@/lib/ai-service';
import { step1Config } from '@/lib/steps';

// 1. Interactive Chat
const chat = await AIService.streamChat({
  messages: [{ role: 'user', content: 'My app idea' }],
  systemPrompt: step1Config.systemPrompt
});

// 2. Document Generation
const doc = await AIService.streamDocument({
  stepName: 'One Pager',
  chatHistory: mySavedHistory
});
```

## Testing
Core logic is tested via Vitest. Run `npm test` to verify the engine integrity.
