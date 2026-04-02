"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { StepConfig, Message } from "@/lib/types";
import { useWizardStore } from "@/app/store";
import ChatInterface from "./ChatInterface";
import ChatInterfaceWithOptions from "./ChatInterfaceWithOptions";
import DocumentPreview from "./DocumentPreview";
import { Terminal, Loader2, X } from 'lucide-react';
import { analytics } from "@/app/utils/analytics";
import { sampleDocs } from "@/lib/utils/sampleDocs";

interface WizardStepProps {
  config: StepConfig;
  stepKey: "onePager" | "devSpec" | "checklist" | "agentsMd";
  onApproveAndNext: () => void;
}

export default function WizardStep({ config, stepKey, onApproveAndNext }: WizardStepProps) {
  const { steps, isGenerating, setIsGenerating, updateStepChat, updateStepDoc, approveStep, resetCounter } =
    useWizardStore();
  const stepData = steps[stepKey];
  const enableMultipleChoice = stepKey === "onePager";

  const [error, setError] = useState<string | null>(null);
  const [showExampleModal, setShowExampleModal] = useState(false);
  const hasGeneratedBefore = useRef(!!stepData.generatedDoc);

  const sampleDocMap: Record<string, string> = {
    onePager: sampleDocs.onePager,
    devSpec: sampleDocs.devSpec,
    checklist: sampleDocs.promptPlan,
    agentsMd: sampleDocs.agentsMd,
  };

  const handleMessagesChange = useCallback((messages: Message[]) => {
    updateStepChat(stepKey, messages);
  }, [stepKey, updateStepChat]);

  // Collect previous documents and chat histories for context
  const documentInputsForChat: Record<string, string> = {};
  const chatContextForChat: Record<string, string> = {};
  
  if (config.documentInputs.length > 0) {
    for (const inputKey of config.documentInputs) {
      const key = inputKey as keyof typeof steps;
      if (steps[key]?.generatedDoc) {
        documentInputsForChat[inputKey] = steps[key].generatedDoc!;
      }
      if (steps[key]?.chatHistory.length > 0) {
        chatContextForChat[inputKey] = steps[key].chatHistory
          .map(m => `${m.role.toUpperCase()}: ${m.content}`)
          .join('\n\n');
      }
    }
  }

  const handleGenerate = useCallback(async () => {
    console.log(`[WizardStep] handleGenerate started for ${stepKey}`);
    setIsGenerating(true);
    setError(null);

    try {
      const documentInputs: Record<string, string> = {};
      const chatContext: Record<string, string> = {};
      
      if (config.documentInputs.length > 0) {
        for (const inputKey of config.documentInputs) {
          const key = inputKey as keyof typeof steps;
          if (steps[key]?.generatedDoc) {
            documentInputs[inputKey] = steps[key].generatedDoc!;
          }
          if (steps[key]?.chatHistory.length > 0) {
            chatContext[inputKey] = steps[key].chatHistory
              .map(m => `${m.role.toUpperCase()}: ${m.content}`)
              .join('\n\n');
          }
        }
      }

      console.log(`[WizardStep] Fetching /api/generate-doc...`);
      const response = await fetch("/api/generate-doc", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          chatHistory: stepData.chatHistory,
          stepName: config.stepName,
          documentInputs,
          chatContext, // Pass chat context here too
          generationPrompt: config.generationPrompt,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: "Unknown error" }));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      if (!response.body) {
        throw new Error("No response body received from API");
      }

      console.log(`[WizardStep] Stream started...`);
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let generatedDoc = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        generatedDoc += decoder.decode(value, { stream: true });
        // Update incrementally if needed, but for now we'll wait for end
      }
      
      // Final flush
      generatedDoc += decoder.decode();

      if (!generatedDoc) {
        throw new Error("API returned an empty document");
      }

      console.log(`[WizardStep] Generation complete, length: ${generatedDoc.length}`);

      // Append attribution for AGENTS.md
      if (stepKey === "agentsMd") {
        generatedDoc = generatedDoc.trimEnd() + "\n\n<!-- Generated with vibescaffold.dev -->\n";
      }

      updateStepDoc(stepKey, generatedDoc);

      // Track successful document generation
      analytics.trackDocumentGenerate(config.stepName, true);

      hasGeneratedBefore.current = true;

      setTimeout(() => {
        const previewElement = document.getElementById('preview-box');
        if (previewElement) {
          previewElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      }, 100);
    } catch (err) {
      console.error(`[WizardStep] handleGenerate error:`, err);
      setError(err instanceof Error ? err.message : "Failed to generate document");
      // Track failed document generation
      analytics.trackDocumentGenerate(config.stepName, false);
    } finally {
      setIsGenerating(false);
    }
  }, [config.documentInputs, config.stepName, stepData.chatHistory, stepKey, steps, updateStepDoc, setIsGenerating, config.generationPrompt]);

  const handleApprove = () => {
    approveStep(stepKey);
  };

  useEffect(() => {
    const handleTriggerGenerate = () => {
      console.log(`[WizardStep] triggerGenerate event received`);
      if (stepData.chatHistory.length > 0 && !isGenerating) {
        handleGenerate();
      } else {
        console.warn(`[WizardStep] Cannot generate: history empty or already generating`, {
          historyLength: stepData.chatHistory.length,
          isGenerating
        });
      }
    };
    window.addEventListener('triggerGenerate', handleTriggerGenerate);
    return () => window.removeEventListener('triggerGenerate', handleTriggerGenerate);
  }, [stepData.chatHistory.length, isGenerating, handleGenerate]);

  return (
    <>
      {error && (
        <div className="bg-red-950/20 border border-red-900/50 p-4 m-6 animate-pulse">
          <div className="flex items-center gap-2 mb-1">
            <X className="w-4 h-4 text-red-500" />
            <p className="text-red-400 font-bold font-mono text-xs uppercase tracking-widest">Generation Error</p>
          </div>
          <p className="text-red-200/70 font-mono text-xs ml-6">{error}</p>
        </div>
      )}

      {/* Chat Box */}
      <div className="flex flex-col flex-1 min-h-[400px]">
        <div className="px-6 py-4 border-b border-zinc-800 bg-zinc-950">
          <div className="text-sm text-[#a1a1aa]">
            {config.userInstructions}
          </div>
          <button
            onClick={() => setShowExampleModal(true)}
            className="text-xs text-accent hover:text-accent-light font-mono mt-2 inline-flex items-center gap-1 transition-all duration-200"
          >
            See example output →
          </button>
        </div>

        <div className="flex-1 overflow-hidden bg-zinc-950">
          {enableMultipleChoice ? (
            <ChatInterfaceWithOptions
              key={`${stepKey}-${resetCounter}`}
              systemPrompt={config.systemPrompt}
              initialMessages={stepData.chatHistory}
              onMessagesChange={handleMessagesChange}
              documentInputs={documentInputsForChat}
              chatContext={chatContextForChat}
              initialGreeting={config.initialGreeting}
              stepName={config.stepName}
              placeholder={config.inputPlaceholder}
              quickStartSuggestions={config.quickStartSuggestions}
            />
          ) : (
            <ChatInterface
              key={`${stepKey}-${resetCounter}`}
              systemPrompt={config.systemPrompt}
              initialMessages={stepData.chatHistory}
              onMessagesChange={handleMessagesChange}
              documentInputs={documentInputsForChat}
              chatContext={chatContextForChat}
              initialGreeting={config.initialGreeting}
              stepName={config.stepName}
              placeholder={config.inputPlaceholder}
              quickStartSuggestions={config.quickStartSuggestions}
            />
          )}
        </div>
      </div>

      {/* Loading Box */}
      {isGenerating && !stepData.generatedDoc && (
        <div id="preview-box" className="border-t border-zinc-800 bg-zinc-950 p-8">
          <div className="flex flex-col items-center justify-center py-16">
            <Loader2 className="w-12 h-12 text-white animate-spin mb-6" />
            <div className="text-sm font-mono font-bold text-white mb-2 tracking-widest">
               GENERATING_ASSETS...
            </div>
            <div className="text-xs text-[#a1a1aa] font-mono">
              Processing {config.stepName} requirements
            </div>

            <div className="w-64 h-1 bg-zinc-800 mt-8 overflow-hidden">
               <div className="h-full bg-zinc-400 w-1/2 animate-[slide_1s_linear_infinite]"></div>
            </div>
          </div>
        </div>
      )}

      {/* Preview Box */}
      {stepData.generatedDoc && (
        <div id="preview-box" className="border-t border-zinc-800 min-h-[500px] h-[70vh] max-h-[800px] flex flex-col">
          <DocumentPreview
            content={stepData.generatedDoc}
            onRegenerate={handleGenerate}
          />
        </div>
      )}

      {error && (
        <div className="bg-red-950/20 border border-red-900/50 p-4 m-6">
          <p className="text-red-400 font-mono text-xs">ERROR: {error}</p>
        </div>
      )}

      {showExampleModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-backdropEnter">
          <div className="bg-zinc-900 border border-zinc-800 max-w-3xl w-full max-h-[80vh] flex flex-col animate-modalEnter shadow-2xl">
            <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-800">
              <div>
                <div className="text-2xs font-mono text-zinc-500 uppercase tracking-widest mb-1">Example Output</div>
                <div className="text-lg font-bold text-white">{config.stepName}</div>
              </div>
              <button
                onClick={() => setShowExampleModal(false)}
                className="text-zinc-400 hover:text-white transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="flex-1 overflow-auto p-6">
              <pre className="text-sm font-mono text-zinc-300 whitespace-pre-wrap leading-relaxed">
                {sampleDocMap[stepKey]}
              </pre>
            </div>
            <div className="px-6 py-4 border-t border-zinc-800 bg-zinc-950">
              <p className="text-xs text-zinc-500">
                This is sample output for a Photo Captioner app. Your output will be customized to your idea.
              </p>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
