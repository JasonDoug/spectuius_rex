import { createOpenAI } from "@ai-sdk/openai";

/**
 * Creates an OpenAI-compatible model instance
 * Supports custom base URLs (e.g., OpenRouter, Ollama) via OPENAI_BASE_URL env var
 */
export function getAIModel(modelName?: string) {
  const baseURL = process.env.OPENAI_BASE_URL;
  const apiKey = process.env.OPENAI_API_KEY;
  
  if (baseURL) {
    console.log(`[AI] Using custom baseURL: ${baseURL}`);
  }
  
  const openaiProvider = createOpenAI({
    apiKey: apiKey || "", 
    baseURL: baseURL || undefined,
    compatibility: "compatible",
  });

  const finalModelName = modelName || process.env.OPENAI_MODEL || "gpt-4o";
  console.log(`[AI] Initializing model: ${finalModelName}`);

  // Use .chat() to force the Chat Completions API instead of the newer Responses API
  // This is required for local providers like Ollama
  return openaiProvider.chat(finalModelName);
}
