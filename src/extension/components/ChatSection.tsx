import React, { useRef, useEffect, useState } from 'react';
import {
  Send,
  MessageSquare,
  Loader2,
  Sparkles,
  Trash2,
  Bot,
  User,
} from 'lucide-react';
import type { ChatMessage } from '../store';

// ── Types ────────────────────────────────────────────────────────────────────

interface ChatSectionProps {
  messages: ChatMessage[];
  onSend: (message: string) => Promise<void>;
  onClear: () => void;
  isLoading: boolean;
  connected: boolean;
  hasProfile: boolean;
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatTime(ts: number): string {
  const d = new Date(ts);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

const SUGGESTION_CHIPS = [
  'Help me answer "Why do you want to work here?"',
  'What are my strongest skills?',
  'Draft an elevator pitch for me',
  'How should I explain a career gap?',
  'What questions should I ask in an interview?',
];

import { marked } from 'marked';

marked.setOptions({
  gfm: true,
  breaks: true,
});

function renderMarkdown(content: string): string {
  try {
    return marked.parse(content) as string;
  } catch {
    return content;
  }
}

const MessageBubble = ({ message }: { message: ChatMessage }) => {
  const isUser = message.role === 'user';

  return (
    <div className={`flex gap-2.5 ${isUser ? 'flex-row-reverse' : 'flex-row'} animate-fade-in`}>
      {/* Avatar */}
      <div
        className={`w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 ${
          isUser
            ? 'bg-blue-500/15 text-blue-400'
            : 'bg-gradient-to-br from-violet-500/20 to-blue-500/20 text-violet-400'
        }`}
      >
        {isUser ? <User className="w-3.5 h-3.5" /> : <Bot className="w-3.5 h-3.5" />}
      </div>

      {/* Bubble */}
      <div
        className={`max-w-[85%] rounded-2xl px-3.5 py-2.5 text-[13px] leading-relaxed ${
          isUser
            ? 'bg-blue-600 text-white rounded-br-md'
            : 'bg-(--bg-tertiary) text-(--text-secondary) border border-(--border-subtle) rounded-bl-md'
        }`}
      >
        {isUser ? (
          <div className="whitespace-pre-wrap break-words">{message.content}</div>
        ) : (
          <div
            className="chat-markdown break-words"
            dangerouslySetInnerHTML={{ __html: renderMarkdown(message.content) }}
          />
        )}
        <div
          className={`text-[10px] mt-1.5 ${
            isUser ? 'text-blue-200/60' : 'text-(--text-tertiary)'
          }`}
        >
          {formatTime(message.timestamp)}
        </div>
      </div>
    </div>
  );
};

// ── Typing Indicator ─────────────────────────────────────────────────────────

const TypingIndicator = () => (
  <div className="flex gap-2.5 animate-fade-in">
    <div className="w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 bg-gradient-to-br from-violet-500/20 to-blue-500/20 text-violet-400">
      <Bot className="w-3.5 h-3.5" />
    </div>
    <div className="bg-(--bg-tertiary) border border-(--border-subtle) rounded-2xl rounded-bl-md px-4 py-3">
      <div className="flex gap-1.5 items-center">
        <div className="w-1.5 h-1.5 rounded-full bg-(--text-tertiary) animate-bounce" style={{ animationDelay: '0ms' }} />
        <div className="w-1.5 h-1.5 rounded-full bg-(--text-tertiary) animate-bounce" style={{ animationDelay: '150ms' }} />
        <div className="w-1.5 h-1.5 rounded-full bg-(--text-tertiary) animate-bounce" style={{ animationDelay: '300ms' }} />
      </div>
    </div>
  </div>
);

// ── Chat Section ─────────────────────────────────────────────────────────────

export const ChatSection = ({
  messages,
  onSend,
  onClear,
  isLoading,
  connected,
  hasProfile,
}: ChatSectionProps) => {
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // Auto-scroll on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isLoading]);

  // Auto-resize textarea
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.style.height = 'auto';
      inputRef.current.style.height = `${Math.min(inputRef.current.scrollHeight, 120)}px`;
    }
  }, [input]);

  const handleSend = async () => {
    const trimmed = input.trim();
    if (!trimmed || isLoading) return;
    setInput('');
    await onSend(trimmed);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleSuggestionClick = (suggestion: string) => {
    if (isLoading) return;
    onSend(suggestion);
  };

  // ── Empty State ──────────────────────────────────────────────────────

  if (!hasProfile) {
    return (
      <div className="card">
        <div className="empty-state">
          <div className="empty-state-icon">
            <MessageSquare className="w-6 h-6" />
          </div>
          <h3 className="empty-state-title">Set up your profile first</h3>
          <p className="empty-state-description">
            The AI assistant needs your profile data to provide personalized answers.
          </p>
        </div>
      </div>
    );
  }

  if (!connected) {
    return (
      <div className="card">
        <div className="empty-state">
          <div className="empty-state-icon">
            <MessageSquare className="w-6 h-6" />
          </div>
          <h3 className="empty-state-title">Server Offline</h3>
          <p className="empty-state-description">
            Connect to the Autoply API server to use the AI assistant.
          </p>
        </div>
      </div>
    );
  }

  // ── Main Chat UI ─────────────────────────────────────────────────────

  return (
    <div className="flex flex-col h-full" style={{ minHeight: 0 }}>
      {/* Header bar */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-violet-500/20 to-blue-500/20 flex items-center justify-center">
            <Sparkles className="w-4 h-4 text-violet-400" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-(--text-primary)">AI Assistant</h3>
            <p className="text-[10px] text-(--text-tertiary)">Powered by your profile</p>
          </div>
        </div>

        {messages.length > 0 && (
          <button
            onClick={onClear}
            className="p-2 rounded-lg hover:bg-(--bg-tertiary) text-(--text-tertiary) hover:text-rose-400 transition-colors"
            aria-label="Clear chat"
            title="Clear conversation"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Messages area */}
      <div className="flex-1 overflow-y-auto space-y-3 pb-3 min-h-[200px] max-h-[calc(100vh-320px)]">
        {messages.length === 0 ? (
          <div className="space-y-4 pt-4">
            {/* Welcome */}
            <div className="text-center space-y-2 py-4">
              <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-violet-500/15 to-blue-500/15 flex items-center justify-center mx-auto">
                <MessageSquare className="w-6 h-6 text-violet-400" />
              </div>
              <h4 className="text-sm font-semibold text-(--text-primary)">
                Ask me anything
              </h4>
              <p className="text-xs text-(--text-tertiary) max-w-[240px] mx-auto leading-relaxed">
                I know your profile, resume, and cover letter. Ask me interview questions, career
                advice, or help with applications.
              </p>
            </div>

            {/* Suggestion chips */}
            <div className="space-y-2">
              {SUGGESTION_CHIPS.map((chip) => (
                <button
                  key={chip}
                  onClick={() => handleSuggestionClick(chip)}
                  className="w-full text-left px-3.5 py-2.5 rounded-xl bg-(--bg-secondary) border border-(--border-subtle) text-xs text-(--text-secondary) hover:bg-(--bg-tertiary) hover:border-(--border-default) hover:text-(--text-primary) transition-all duration-200"
                >
                  {chip}
                </button>
              ))}
            </div>
          </div>
        ) : (
          <>
            {messages.map((msg) => (
              <MessageBubble key={msg.id} message={msg} />
            ))}
            {isLoading && <TypingIndicator />}
          </>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input area */}
      <div className="pt-2 border-t border-(--border-subtle)">
        <div className="flex items-end gap-2 bg-(--bg-secondary) border border-(--border-subtle) rounded-2xl px-3 py-2 focus-within:border-(--border-focus) transition-colors">
          <textarea
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask anything…"
            rows={1}
            className="flex-1 bg-transparent text-sm text-(--text-primary) placeholder:text-(--text-tertiary) outline-none resize-none"
            style={{ maxHeight: '120px' }}
            disabled={isLoading}
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || isLoading}
            className="p-2 rounded-xl bg-blue-600 text-white hover:bg-blue-500 disabled:opacity-30 disabled:hover:bg-blue-600 transition-all flex-shrink-0"
            aria-label="Send message"
          >
            {isLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Send className="w-4 h-4" />
            )}
          </button>
        </div>
        <p className="text-[10px] text-(--text-tertiary) text-center mt-1.5">
          Shift+Enter for new line
        </p>
      </div>
    </div>
  );
};
