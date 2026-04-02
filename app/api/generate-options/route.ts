import { getAIModel } from "@/lib/utils/ai";
import { generateObject } from "ai";

import { questionOptionsSchema } from "@/lib/schemas/questionOptions";

export const runtime = "edge";

const systemPrompt = [
  "You generate multiple-choice options for a clarifying question in a product requirements chat.",
  "",
  "Hard rules:",
  "- reasoning MUST come first and be a brief analysis before committing to options.",
  "- options MUST be 3-6 distinct, specific, concrete options that are plausible answers to the question.",
  '- Do not include generic options like "Other", "None of the above", or vague placeholders.',
  '- Never include: "Other", "None of the above", "Something else".',
  "- recommendedIndex MUST be a 0-based index into options or null if no safe recommendation.",
  '- confidence MUST be one of: "weak" | "medium" | "strong".',
].join("\n");

function summarizeChat(history: any[]): string {
  if (!history || !Array.isArray(history)) return "(none)";
  return history
    .slice(-8)
    .map(m => `${m.role === "user" ? "User" : "Assistant"}: ${m.content.substring(0, 200)}`)
    .join("\n");
}

export async function POST(req: Request) {
  const startTime = Date.now();
  console.log(`[API] Received generate-options request at ${new Date().toISOString()}`);
  
  try {
    const body = await req.json();
    const questionText = body?.questionText;
    let conversationSummary = body?.conversationSummary;

    // Handle TUI providing chatHistory instead of a summary
    if (!conversationSummary && body?.chatHistory) {
      console.log("[API] Summarizing chatHistory provided by client...");
      conversationSummary = summarizeChat(body.chatHistory);
    }

    console.log(`[API] QuestionText: ${questionText?.substring(0, 100)}...`);

    if (typeof questionText !== "string" || questionText.trim().length === 0) {
      console.error("[API] questionText is required");
      return new Response("questionText is required", { status: 400 });
    }

    const modelName =
      process.env.OPENAI_OPTIONS_MODEL || process.env.OPENAI_MODEL || "gpt-4o";

    console.log(`[API] Using model: ${modelName}`);

    const prompt = [
      "CONVERSATION SUMMARY:",
      typeof conversationSummary === "string" && conversationSummary.trim().length > 0
        ? conversationSummary.trim()
        : "(none)",
      "",
      "QUESTION:",
      questionText.trim(),
      "",
      "IMPORTANT: Return ONLY the raw JSON object. Do not include markdown blocks or any other text.",
    ].join("\n");

    const { object } = await generateObject({
      model: getAIModel(modelName),
      messages: [
        { role: "system", content: systemPrompt },
        { role: "user", content: prompt }
      ],
      schema: questionOptionsSchema,
    });

    console.log(`[API] Options generated successfully in ${Date.now() - startTime}ms`);
    return new Response(JSON.stringify(object), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  } catch (error) {
    console.error("[API] Generate options API error:", error);
    if (error instanceof Error) {
      console.error(`[API] Error details: ${error.message}\n${error.stack}`);
    }
    return new Response(
      JSON.stringify({ error: "Failed to generate options" }),
      {
        status: 500,
        headers: { "Content-Type": "application/json" },
      }
    );
  }
}
