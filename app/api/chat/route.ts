import { AIService } from "@/lib/ai-service";

export const runtime = "edge";

export async function POST(req: Request) {
  const startTime = Date.now();
  console.log(`[API] Received chat request at ${new Date().toISOString()}`);
  
  try {
    const body = await req.json();
    const { messages, systemPrompt, documentInputs, chatContext, stream = true } = body;

    if (!messages || !Array.isArray(messages)) {
      console.error("[API] Invalid messages format received");
      return new Response("Invalid messages format", { status: 400 });
    }

    console.log(`[API] Chat request: ${messages.length} messages (stream: ${stream})`);
    console.log(`[API] System Prompt length: ${systemPrompt?.length || 0}`);

    const result = await AIService.streamChat({
      messages,
      systemPrompt,
      documentInputs,
      chatContext,
      stream,
    });

    console.log(`[API] AIService returned result in ${Date.now() - startTime}ms`);

    if (stream) {
      console.log("[API] Returning stream response...");
      return (result as any).toTextStreamResponse();
    }

    console.log("[API] Returning static text response...");
    return new Response((result as any).text, {
      status: 200,
      headers: { "Content-Type": "text/plain; charset=utf-8" },
    });
  } catch (error) {
    console.error("[API] Chat API error:", error);
    if (error instanceof Error) {
      console.error(`[API] Error details: ${error.message}\n${error.stack}`);
    }

    return new Response(
      JSON.stringify({ error: "Failed to process chat request" }),
      {
        status: 500,
        headers: { "Content-Type": "application/json" },
      }
    );
  }
}
