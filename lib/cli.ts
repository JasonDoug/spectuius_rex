import { AIService } from "./ai-service";
import { step1Config } from "./steps/step1-config";
import * as readline from "node:readline/promises";

/**
 * A simple CLI demonstration of the decoupled AIService.
 * Run with: npx tsx lib/cli.ts "My app idea"
 */
async function main() {
  const idea = process.argv[2];
  if (!idea) {
    console.log("Usage: npx tsx lib/cli.ts \"Your app idea\"");
    process.exit(1);
  }

  console.log("\n🚀 Vibe Scaffold CLI (Alpha)");
  console.log("============================");
  console.log(`Idea: ${idea}\n`);

  console.log("⏳ Generating One Pager...\n");

  const result = await AIService.streamDocument({
    stepName: step1Config.stepName,
    chatHistory: [{ role: "user", content: idea }],
    generationPrompt: step1Config.generationPrompt,
  });

  process.stdout.write("📄 ONE PAGER:\n\n");
  for await (const chunk of result.textStream) {
    process.stdout.write(chunk);
  }
  process.stdout.write("\n\n✅ Done.\n");
}

if (require.main === module) {
  main().catch(console.error);
}
