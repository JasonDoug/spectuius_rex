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
  
  if (!apiKey) {
    console.warn("[AI] Warning: OPENAI_API_KEY is not set");
  } else {
    console.log(`[AI] API Key found (length: ${apiKey.length})`);
  }

  const openaiProvider = createOpenAI({
    apiKey: apiKey || "", // Provide empty string as fallback to avoid LoadAPIKeyError if baseURL handles it
    baseURL: baseURL || undefined,
  });

  const finalModelName = modelName || process.env.OPENAI_MODEL || "gpt-4o";
  console.log(`[AI] Initializing model: ${finalModelName}`);

  return openaiProvider(finalModelName);
}
