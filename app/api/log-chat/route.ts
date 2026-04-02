import { sql } from "@vercel/postgres";
import { NextRequest, NextResponse } from "next/server";

interface LogChatBody {
  clientId: string;
  sessionId: string;
  stepName?: string;
  role: "user" | "assistant";
  content: string;
}

export async function POST(request: NextRequest) {
  try {
    const { clientId, sessionId, stepName, role, content }: LogChatBody = await request.json();

    if (!clientId || !sessionId || !role || !content) {
      return NextResponse.json({ error: "Missing required fields" }, { status: 400 });
    }

    // Check if database is configured to avoid loud console errors in local dev
    if (!process.env.POSTGRES_URL) {
      return NextResponse.json({ success: true, warning: "Database not configured" });
    }

    await sql`
      INSERT INTO chat_logs (client_id, session_id, step_name, role, content)
      VALUES (${clientId}, ${sessionId}, ${stepName || null}, ${role}, ${content})
    `;

    return NextResponse.json({ success: true });
  } catch (error) {
    // Only log actual errors, not missing connection strings
    if (error instanceof Error && !error.message.includes('missing_connection_string')) {
      console.error("Failed to log chat message:", error);
    }
    // Fire-and-forget pattern: return success anyway
    return NextResponse.json({ success: false });
  }
}
