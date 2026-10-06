import { Suspense } from "react";
import { AuthCard } from "@/components/auth/AuthCard";
import LoginForm from "@/components/auth/LoginForm";

// Force dynamic rendering to prevent prerender errors during `next build`
export const dynamic = "force-dynamic";

export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <AuthCard
        title="Welcome back"
        description="Sign in to your CyberSaaS management account."
      >
        <Suspense fallback={<div className="text-sm text-slate-500 text-center">Loading form...</div>}>
          <LoginForm role="platform_admin" />
        </Suspense>
      </AuthCard>
    </main>
  );
}