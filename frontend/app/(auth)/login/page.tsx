"use client";

import Link from "next/link";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const loginOptions = [
  {
    title: "SaaS Owner",
    description: "Manage the CyberSaaS platform.",
    href: "/login/saas-owner",
    button: "Sign in as SaaS Owner",
  },
  {
    title: "Cyber Owner",
    description: "Manage your cyber café business.",
    href: "/login/owner",
    button: "Sign in as Cyber Owner",
  },
  {
    title: "Cyber Attendant",
    description: "Operate your assigned cyber café branch.",
    href: "/login/cyber-attendant",
    button: "Sign in as Attendant",
  },
];

export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-white px-4 py-10">
      <div className="w-full max-w-5xl">
        <div className="mb-10 text-center">
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">
            Welcome to CyberSaaS
          </h1>

          <p className="mt-3 text-slate-500">
            Choose how you want to sign in.
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-3">
          {loginOptions.map((option) => (
            <Card
              key={option.href}
              className="border-slate-200 shadow-sm transition-shadow hover:shadow-md"
            >
              <CardHeader>
                <CardTitle className="text-xl text-slate-900">
                  {option.title}
                </CardTitle>

                <CardDescription className="min-h-10">
                  {option.description}
                </CardDescription>
              </CardHeader>

              <CardContent>
                <Link
                  href={option.href}
                  className="inline-flex h-12 w-full items-center justify-center rounded-xl bg-blue-600 px-4 text-sm font-semibold text-white transition-colors hover:bg-blue-700"
                >
                  {option.button}
                </Link>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </main>
  );
}