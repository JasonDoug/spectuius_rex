import { getAIModel } from "@/lib/utils/ai";
import { streamText, generateText, CoreMessage } from "ai";
import { Message } from "@/lib/types";

/**
 * Core AI Service that handles prompt construction and model interaction.
 * This is designed to be independent of the transport layer (HTTP, CLI, etc.)
 */
export const AIService = {
  /**
   * Generates a technical document based on chat history and context
   */
  async streamDocument(params: {
    stepName: string;
    chatHistory: Message[];
    documentInputs?: Record<string, string>;
    chatContext?: Record<string, string>;
    generationPrompt?: string;
  }) {
    const { stepName, chatHistory, documentInputs, chatContext, generationPrompt: customPrompt } = params;

    // Build generation prompt
    let prompt = customPrompt ? `${customPrompt}\n\n` : "";

    // Format conversation history
    const conversationHistory = chatHistory
      .map((msg) => `${msg.role === "user" ? "User" : "Assistant"}: ${msg.content}`)
      .join("\n\n");

    // Add previous documents context
    if (documentInputs && Object.keys(documentInputs).length > 0) {
      prompt += "You also have access to these previously generated documents for context:\n\n";
      for (const [key, value] of Object.entries(documentInputs)) {
        prompt += `### ${key}:\n${value}\n\n`;
      }
    }

    // Add previous chat history context
    if (chatContext && Object.keys(chatContext).length > 0) {
      prompt += "You also have access to the conversation history from previous steps for context:\n\n";
      for (const [key, value] of Object.entries(chatContext)) {
        prompt += `### ${key} chat history:\n${value}\n\n`;
      }
    }

    prompt += `Based on the conversation history below, generate a comprehensive ${stepName} document in markdown format.\n\nConversation history:\n${conversationHistory}\n\nPlease generate the ${stepName} document now in markdown format:`;

    const modelName = process.env.OPENAI_MODEL || "gpt-4o";
    const model = getAIModel(modelName);

    return streamText({
      model,
      prompt,
    });
  },

  async streamChat(params: {
    messages: Message[];
    systemPrompt?: string;
    documentInputs?: Record<string, string>;
    chatContext?: Record<string, string>;
    stream?: boolean;
  }) {
    const { messages, systemPrompt, documentInputs, chatContext, stream = true } = params;

    let fullSystemPrompt = systemPrompt || "You are a helpful assistant.";

    if (documentInputs && Object.keys(documentInputs).length > 0) {
      fullSystemPrompt += "\n\n--- PREVIOUS DOCUMENTS FOR CONTEXT ---\n\n";
      for (const [key, value] of Object.entries(documentInputs)) {
        fullSystemPrompt += `### ${key}:\n${value}\n\n`;
      }
      fullSystemPrompt += "--- END OF PREVIOUS DOCUMENTS ---\n";
    }

    if (chatContext && Object.keys(chatContext).length > 0) {
      fullSystemPrompt += "\n\n--- PREVIOUS CHAT HISTORY FOR CONTEXT ---\n\n";
      for (const [key, value] of Object.entries(chatContext)) {
        fullSystemPrompt += `### ${key} chat history:\n${value}\n\n`;
      }
      fullSystemPrompt += "--- END OF PREVIOUS CHAT HISTORY ---\n";
    }

    const modelName = process.env.OPENAI_MODEL || "gpt-4o";
    const model = getAIModel(modelName);

    // Prepend system prompt as a message for better compatibility with some providers
    const coreMessages: CoreMessage[] = [
      { role: "system", content: fullSystemPrompt },
      ...messages.map(m => ({ role: m.role, content: m.content } as CoreMessage))
    ];

    if (stream) {
      return streamText({
        model,
        messages: coreMessages,
      });
    }

    return generateText({
      model,
      messages: coreMessages,
    });
  }
};
