import React from 'react';
import {
  Sparkles,
  Layers,
  Check,
  ArrowRight,
  UserCheck,
  AlertCircle,
} from 'lucide-react';
import type { Application, Profile } from '../../types';

interface HomeInsightsSectionProps {
  currentTabUrl?: string;
  applications: Application[];
  profile: Profile | null;
  onNavigateTab: (tab: 'analytics' | 'profile') => void;
}

export const HomeInsightsSection: React.FC<HomeInsightsSectionProps> = ({
  applications,
  profile,
  onNavigateTab,
}) => {
  // 1. Application Status Funnel Stats
  const submittedCount = applications.filter((a) => a.status === 'submitted').length;
  const reviewRequiredCount = applications.filter((a) => a.status === 'review_required' || a.status === 'pending').length;
  const filledCount = applications.filter((a) => a.status === 'filled').length;
  const failedCount = applications.filter((a) => a.status === 'failed').length;

  // 2. Profile Readiness Calculations
  const profileItems = [
    { label: 'Full Name & Email', ready: Boolean(profile?.name && profile?.email) },
    { label: 'Phone & Location', ready: Boolean(profile?.phone && profile?.location) },
    { label: 'Base Resume Uploaded', ready: Boolean(profile?.base_resume) },
    { label: 'LinkedIn / Portfolio URL', ready: Boolean(profile?.linkedin_url || profile?.portfolio_url) },
    { label: 'Work Experience Logged', ready: Boolean(profile?.experience && profile.experience.length > 0) },
  ];

  const readyCount = profileItems.filter((item) => item.ready).length;
  const readinessPercent = Math.round((readyCount / profileItems.length) * 100);

  return (
    <div className="space-y-4 animate-fade-in">

      {/* ── 2. Application Funnel & Status Summary ───────────────────────── */}
      <div className="surface p-4 space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Layers className="w-4 h-4 text-blue-400" />
            <h3 className="text-xs font-bold uppercase tracking-wider text-zinc-300">
              Application Pipeline
            </h3>
          </div>
          <button
            onClick={() => onNavigateTab('analytics')}
            className="text-[0.7rem] font-medium text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-1"
          >
            <span>View All ({applications.length})</span>
            <ArrowRight className="w-3 h-3" />
          </button>
        </div>

        <div className="grid grid-cols-3 gap-2">
          <div className="p-2.5 rounded-lg bg-zinc-900/60 border border-zinc-800 text-center">
            <span className="block text-[0.65rem] font-semibold uppercase text-amber-400/90">
              Review Needed
            </span>
            <span className="text-lg font-extrabold text-zinc-100">{reviewRequiredCount}</span>
          </div>

          <div className="p-2.5 rounded-lg bg-zinc-900/60 border border-zinc-800 text-center">
            <span className="block text-[0.65rem] font-semibold uppercase text-blue-400">
              Ready / Filled
            </span>
            <span className="text-lg font-extrabold text-zinc-100">{filledCount}</span>
          </div>

          <div className="p-2.5 rounded-lg bg-zinc-900/60 border border-zinc-800 text-center">
            <span className="block text-[0.65rem] font-semibold uppercase text-emerald-400">
              Submitted
            </span>
            <span className="text-lg font-extrabold text-zinc-100">{submittedCount}</span>
          </div>
        </div>

        {/* Visual Progress Ratio Bar */}
        {applications.length > 0 && (
          <div className="space-y-1.5 pt-1">
            <div className="flex justify-between text-[0.65rem] text-zinc-400 font-medium">
              <span>Pipeline Success Rate</span>
              <span>
                {Math.round(((submittedCount + filledCount) / (applications.length || 1)) * 100)}%
              </span>
            </div>
            <div className="h-2 w-full bg-zinc-800 rounded-full overflow-hidden flex">
              <div
                className="bg-emerald-500 transition-all duration-500"
                style={{ width: `${(submittedCount / (applications.length || 1)) * 100}%` }}
                title="Submitted"
              />
              <div
                className="bg-blue-500 transition-all duration-500"
                style={{ width: `${(filledCount / (applications.length || 1)) * 100}%` }}
                title="Filled"
              />
              <div
                className="bg-amber-500 transition-all duration-500"
                style={{ width: `${(reviewRequiredCount / (applications.length || 1)) * 100}%` }}
                title="Review Required"
              />
              <div
                className="bg-rose-500 transition-all duration-500"
                style={{ width: `${(failedCount / (applications.length || 1)) * 100}%` }}
                title="Failed"
              />
            </div>
          </div>
        )}
      </div>

      {/* ── 3. Profile Readiness & AI Optimization Score ────────────────────── */}
      <div className="surface p-4 space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <UserCheck className="w-4 h-4 text-emerald-400" />
            <h3 className="text-xs font-bold uppercase tracking-wider text-zinc-300">
              AI Profile Readiness
            </h3>
          </div>
          <span className="text-xs font-bold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">
            {readinessPercent}% Ready
          </span>
        </div>

        {/* Readiness Checklist */}
        <div className="space-y-1.5 pt-1">
          {profileItems.map((item, idx) => (
            <div
              key={idx}
              className="flex items-center justify-between py-1 text-xs border-b border-zinc-800/40 last:border-0"
            >
              <span className="text-zinc-300 font-medium">{item.label}</span>
              {item.ready ? (
                <span className="flex items-center gap-1 text-emerald-400 text-[0.7rem] font-medium">
                  <Check className="w-3.5 h-3.5" />
                  Ready
                </span>
              ) : (
                <span className="flex items-center gap-1 text-amber-400 text-[0.7rem] font-medium">
                  <AlertCircle className="w-3.5 h-3.5" />
                  Missing
                </span>
              )}
            </div>
          ))}
        </div>

        {readinessPercent < 100 && (
          <button
            onClick={() => onNavigateTab('profile')}
            className="w-full mt-2 py-1.5 px-3 rounded-lg bg-zinc-800/80 hover:bg-zinc-800 text-blue-400 text-xs font-semibold flex items-center justify-center gap-1.5 transition-colors border border-zinc-700/50"
          >
            <Sparkles className="w-3.5 h-3.5 text-blue-400" />
            <span>Complete Profile for Higher Autofill Accuracy</span>
          </button>
        )}
      </div>
    </div>
  );
};
