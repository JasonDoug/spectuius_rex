"use client";

import { useState, useEffect } from "react";
import { Star } from "lucide-react";

interface GitHubStarsProps {
  repo: string;
}

export default function GitHubStars({ repo }: GitHubStarsProps) {
  const [stars, setStars] = useState<number | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    fetch(`https://api.github.com/repos/${repo}`)
      .then((res) => {
        if (!res.ok) throw new Error("GitHub API error");
        return res.json();
      })
      .then((data) => {
        if (typeof data.stargazers_count === "number") {
          setStars(data.stargazers_count);
        }
      })
      .catch(() => {
        // Keep stars as null on error
      });
  }, [repo]);

  if (!mounted) {
    return (
      <div className="flex items-center gap-1.5 px-2.5 py-1.5 border border-zinc-700 bg-zinc-900 text-2xs font-mono text-[#a1a1aa]">
        <Star className="w-3 h-3" />
        <span>-</span>
        <span>GitHub</span>
      </div>
    );
  }

  return (
    <a
      href={`https://github.com/${repo}`}
      target="_blank"
      rel="noopener noreferrer"
      className="flex items-center gap-1.5 px-2.5 py-1.5 border border-zinc-700 hover:border-accent bg-zinc-900 text-2xs font-mono text-[#a1a1aa] hover:text-accent transition-colors"
    >
      <Star className="w-3 h-3" />
      <span>{stars !== null ? stars.toLocaleString() : "-"}</span>
      <span>GitHub</span>
    </a>
  );
}
