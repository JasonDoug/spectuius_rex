import { AIService } from "@/lib/ai-service";

export const runtime = "edge";

export async function POST(req: Request) {
  const startTime = Date.now();
  let stepName: string | undefined;

  try {
    const body = await req.json();
    stepName = body.stepName;
    const { chatHistory, documentInputs, chatContext, generationPrompt } = body;

    if (!chatHistory || !Array.isArray(chatHistory)) {
      return new Response("Invalid chat history format", { status: 400 });
    }

    if (!stepName) {
      return new Response("Step name is required", { status: 400 });
    }

    // Log the request details for debugging
    console.log(`[API] Generating document: ${stepName} (${chatHistory.length} messages)`);

    const result = await AIService.streamDocument({
      stepName,
      chatHistory,
      documentInputs,
      chatContext,
      generationPrompt,
    });

    return result.toTextStreamResponse();
  } catch (error) {
    console.error("Generate doc API error:", error);

    return new Response(
      JSON.stringify({ error: "Failed to generate document" }),
      {
        status: 500,
        headers: { "Content-Type": "application/json" },
      }
    );
  }
}
