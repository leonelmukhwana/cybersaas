"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  Check,
  ChevronRight,
  Download,
  FileText,
  Mail,
  MessageCircle,
  ShieldCheck,
  Smartphone,
  WifiOff,
} from "lucide-react";

import { subscriptionService } from "@/services/subscription.service";
import type { SubscriptionPlan } from "@/types/subscription";

const features = [
  {
    title: "Manage Your Cyber",
    description:
      "Manage branches, computers, attendants, customers, services, sessions and daily operations from one place.",
    icon: "⌘",
  },
  {
    title: "Fast Billing & Receipts",
    description:
      "Record computer sessions, printing, photocopying and other services while keeping payment and receipt records organized.",
    icon: "₵",
  },
  {
    title: "Work Even When Internet Is Down",
    description:
      "Cyber Attendants can continue essential operations when internet connectivity is temporarily unavailable, then synchronize when connection returns.",
    icon: "↯",
  },
  {
    title: "Multiple Branches",
    description:
      "Manage one cyber café or multiple branches from a single Cyber Owner account.",
    icon: "⌂",
  },
  {
    title: "Secure Role-Based Access",
    description:
      "Cyber Owners and Attendants get the access they need while sensitive business information stays protected.",
    icon: "✓",
  },
  {
    title: "Reports & Business Visibility",
    description:
      "Understand collections, sales, sessions, expenses and branch performance without complicated spreadsheets.",
    icon: "↗",
  },
];

const complianceFeatures = [
  "Customer identification and registration",
  "Terminal identification",
  "Session start and end records",
  "Payment and receipt records",
  "Branch-level operational records",
  "Audit-friendly reporting",
  "No customer browsing history storage",
  "Offline-first attendant operations",
];

const installationSteps = [
  {
    number: "01",
    title: "Download CyberSaaS Terminal",
    description:
      "Download the official CyberSaaS Terminal Windows installer for the computer or terminal you want to set up.",
  },
  {
    number: "02",
    title: "Install the application",
    description:
      "Open the downloaded installer and follow the installation instructions on your Windows computer.",
  },
  {
    number: "03",
    title: "Register the terminal",
    description:
      "Use the terminal registration information provided from your Cyber Owner dashboard.",
  },
  {
    number: "04",
    title: "Start serving customers",
    description:
      "Once the terminal is registered, the attendant can begin managing sessions and cyber services.",
  },
];

function getWhatsAppUrl() {
  return "https://wa.me/254793750450";
}

export default function Home() {
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [plansLoading, setPlansLoading] = useState(true);
  const [plansError, setPlansError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    async function loadPlans() {
      try {
        setPlansLoading(true);
        setPlansError(null);

        const data = await subscriptionService.getPlans("");

        if (!mounted) return;

        const activePlans = data
          .filter((plan) => plan.is_active)
          .sort((a, b) => {
            if (a.is_lifetime !== b.is_lifetime) {
              return a.is_lifetime ? 1 : -1;
            }

            return Number(a.monthly_price) - Number(b.monthly_price);
          });

        setPlans(activePlans);
      } catch (error) {
        if (!mounted) return;

        console.error("Failed to load subscription packages:", error);
        setPlansError("Unable to load packages right now.");
      } finally {
        if (mounted) {
          setPlansLoading(false);
        }
      }
    }

    loadPlans();

    return () => {
      mounted = false;
    };
  }, []);

  return (
    <main className="min-h-screen bg-white text-slate-900">
      {/* NAVBAR */}
      <header className="sticky top-0 z-50 border-b border-slate-200/80 bg-white/95 backdrop-blur">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 lg:px-8">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-[#0757B8] text-sm font-bold text-white shadow-sm">
              C
            </div>

            <div>
              <div className="text-lg font-bold tracking-tight text-slate-950">
                Cyber<span className="text-[#0757B8]">SaaS</span>
              </div>

              <div className="text-[10px] font-medium uppercase tracking-wider text-slate-400">
                by EPOA
              </div>
            </div>
          </Link>

          <nav className="hidden items-center gap-7 md:flex">
            <Link
              href="#features"
              className="text-sm font-medium text-slate-600 transition hover:text-[#0757B8]"
            >
              Features
            </Link>

            <Link
              href="#compliance"
              className="text-sm font-medium text-slate-600 transition hover:text-[#0757B8]"
            >
              CA Compliance
            </Link>

            <Link
              href="#download"
              className="text-sm font-medium text-slate-600 transition hover:text-[#0757B8]"
            >
              Download
            </Link>

            <Link
              href="#pricing"
              className="text-sm font-medium text-slate-600 transition hover:text-[#0757B8]"
            >
              Packages
            </Link>

            <Link
              href="#contact"
              className="text-sm font-medium text-slate-600 transition hover:text-[#0757B8]"
            >
              Contact
            </Link>
          </nav>

          <div className="flex items-center gap-2.5">
            <Link
              href="/login"
              className="hidden rounded-lg px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-100 sm:block"
            >
              Login
            </Link>

            <Link
              href="/register/owner"
              className="rounded-lg bg-[#0757B8] px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-[#064A9D]"
            >
              Get Started
            </Link>
          </div>
        </div>
      </header>

      {/* HERO */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,rgba(7,87,184,0.12),transparent_36%)]" />

        <div className="mx-auto grid max-w-7xl items-center gap-14 px-5 pb-20 pt-16 lg:grid-cols-2 lg:px-8 lg:pb-28 lg:pt-24">
          <div>
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-blue-100 bg-blue-50 px-3.5 py-1.5 text-sm font-semibold text-[#0757B8]">
              <span className="h-2 w-2 rounded-full bg-[#0757B8]" />
              🇰🇪 Kenyan-built software for Kenyan cyber cafés
            </div>

            <h1 className="max-w-3xl text-4xl font-extrabold leading-[1.08] tracking-tight text-slate-950 sm:text-5xl lg:text-6xl">
              Run your cyber café
              <span className="block text-[#0757B8]">
                the smarter Kenyan way.
              </span>
            </h1>

            <p className="mt-6 max-w-2xl text-lg leading-8 text-slate-600">
              CyberSaaS by EPOA gives Kenyan cyber café owners one platform to
              manage customers, computers, attendants, sessions, printing,
              photocopying, payments, receipts, branches and reports.
            </p>

            <div className="mt-6 flex flex-wrap gap-3 text-sm font-medium text-slate-600">
              <span className="rounded-full bg-slate-100 px-3 py-1.5">
                🇰🇪 Made for Kenya
              </span>

              <span className="rounded-full bg-slate-100 px-3 py-1.5">
                KSh billing
              </span>

              <span className="rounded-full bg-slate-100 px-3 py-1.5">
                M-Pesa ready
              </span>

              <span className="rounded-full bg-slate-100 px-3 py-1.5">
                Offline-ready
              </span>
            </div>

            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Link
                href="/register/owner"
                className="inline-flex items-center justify-center rounded-xl bg-[#0757B8] px-6 py-3.5 text-sm font-bold text-white shadow-lg shadow-blue-900/10 transition hover:bg-[#064A9D]"
              >
                Start Your Cyber
                <ChevronRight className="ml-2 h-4 w-4" />
              </Link>

              <Link
                href="#download"
                className="inline-flex items-center justify-center rounded-xl border border-slate-300 bg-white px-6 py-3.5 text-sm font-bold text-slate-700 transition hover:border-[#0757B8] hover:text-[#0757B8]"
              >
                <Download className="mr-2 h-4 w-4" />
                Download Software
              </Link>
            </div>

            <div className="mt-8 flex flex-wrap gap-x-6 gap-y-3 text-sm text-slate-500">
              <span className="flex items-center gap-2">
                <Check className="h-4 w-4 text-[#0757B8]" />
                Multiple branches
              </span>

              <span className="flex items-center gap-2">
                <Check className="h-4 w-4 text-[#0757B8]" />
                Offline-ready operations
              </span>

              <span className="flex items-center gap-2">
                <Check className="h-4 w-4 text-[#0757B8]" />
                Receipts & reports
              </span>
            </div>

            {/* CONTACT STRIP */}
            <div className="mt-8 flex flex-col gap-3 border-t border-slate-200 pt-6 sm:flex-row sm:items-center">
              <a
                href={getWhatsAppUrl()}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 text-sm font-semibold text-slate-700 transition hover:text-[#0757B8]"
              >
                <MessageCircle className="h-4 w-4 text-[#0757B8]" />
                WhatsApp: 0793 750 450
              </a>

              <a
                href="mailto:epoa2026@gmail.com"
                className="inline-flex items-center gap-2 text-sm font-semibold text-slate-700 transition hover:text-[#0757B8]"
              >
                <Mail className="h-4 w-4 text-[#0757B8]" />
                epoa2026@gmail.com
              </a>
            </div>
          </div>

          {/* HERO DASHBOARD PREVIEW */}
          <div className="relative">
            <div className="absolute -inset-5 rounded-[2rem] bg-blue-100/50 blur-2xl" />

            <div className="relative overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-2xl shadow-slate-900/10">
              <div className="flex h-12 items-center gap-2 border-b border-slate-200 px-4">
                <span className="h-2.5 w-2.5 rounded-full bg-slate-300" />
                <span className="h-2.5 w-2.5 rounded-full bg-slate-300" />
                <span className="h-2.5 w-2.5 rounded-full bg-slate-300" />

                <div className="ml-4 h-6 flex-1 rounded-md bg-slate-100" />
              </div>

              <div className="grid grid-cols-[130px_1fr]">
                <aside className="hidden min-h-[390px] border-r border-slate-200 bg-slate-50 p-4 sm:block">
                  <div className="mb-7 text-sm font-bold text-[#0757B8]">
                    CyberSaaS
                  </div>

                  <div className="space-y-2">
                    {[
                      "Overview",
                      "Branches",
                      "Computers",
                      "Customers",
                      "Reports",
                    ].map((item, index) => (
                      <div
                        key={item}
                        className={`rounded-lg px-3 py-2 text-xs font-medium ${
                          index === 0
                            ? "bg-blue-100 text-[#0757B8]"
                            : "text-slate-500"
                        }`}
                      >
                        {item}
                      </div>
                    ))}
                  </div>
                </aside>

                <div className="p-5">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-xs text-slate-400">
                        Good morning
                      </p>

                      <h3 className="mt-1 text-lg font-bold text-slate-900">
                        Cyber Owner
                      </h3>
                    </div>

                    <div className="flex h-9 w-9 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-[#0757B8]">
                      CO
                    </div>
                  </div>

                  <div className="mt-6 grid grid-cols-2 gap-3">
                    {[
                      ["Active Computers", "24"],
                      ["Branches", "3"],
                      ["Today's Sales", "KES 8,450"],
                      ["Active Sessions", "17"],
                    ].map(([label, value]) => (
                      <div
                        key={label}
                        className="rounded-xl border border-slate-200 bg-white p-3"
                      >
                        <p className="text-[10px] text-slate-400">
                          {label}
                        </p>

                        <p className="mt-1 text-sm font-bold text-slate-900">
                          {value}
                        </p>
                      </div>
                    ))}
                  </div>

                  <div className="mt-4 rounded-xl border border-slate-200 p-4">
                    <div className="flex items-center justify-between">
                      <p className="text-xs font-bold text-slate-700">
                        Today's activity
                      </p>

                      <span className="text-[10px] text-[#0757B8]">
                        View report
                      </span>
                    </div>

                    <div className="mt-5 flex h-24 items-end gap-2">
                      {[35, 55, 42, 75, 58, 82, 68, 92, 72, 88, 76, 96].map(
                        (height, index) => (
                          <div
                            key={index}
                            className="flex-1 rounded-t-md bg-[#0757B8]/80"
                            style={{ height: `${height}%` }}
                          />
                        )
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* FLOATING TRUST CARD */}
            <div className="absolute -bottom-5 -left-5 hidden rounded-2xl border border-slate-200 bg-white p-4 shadow-xl sm:block">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
                  <ShieldCheck className="h-5 w-5" />
                </div>

                <div>
                  <p className="text-xs font-bold text-slate-900">
                    Compliance-focused
                  </p>

                  <p className="mt-0.5 text-[11px] text-slate-500">
                    Built around Kenyan cyber café needs
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* KENYAN TRUST BAR */}
      <section className="border-y border-slate-200 bg-slate-50">
        <div className="mx-auto grid max-w-7xl gap-6 px-5 py-8 sm:grid-cols-3 lg:px-8">
          <div className="flex items-center gap-4">
            <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
              🇰🇪
            </div>

            <div>
              <p className="font-bold text-slate-950">Built for Kenya</p>
              <p className="text-sm text-slate-500">
                Kenyan cyber café workflows
              </p>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
              <Smartphone className="h-5 w-5" />
            </div>

            <div>
              <p className="font-bold text-slate-950">M-Pesa Ready</p>
              <p className="text-sm text-slate-500">
                Designed for Kenyan payments
              </p>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-blue-50 text-[#0757B8]">
              <WifiOff className="h-5 w-5" />
            </div>

            <div>
              <p className="font-bold text-slate-950">Offline First</p>
              <p className="text-sm text-slate-500">
                Keep essential work moving
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* FEATURES */}
      <section id="features" className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
          <div className="mx-auto max-w-2xl text-center">
            <p className="text-sm font-bold uppercase tracking-wider text-[#0757B8]">
              Everything in one place
            </p>

            <h2 className="mt-3 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">
              Built for the way Kenyan cyber cafés actually work
            </h2>

            <p className="mt-4 text-slate-600">
              From computer sessions to photocopying and receipts, CyberSaaS
              brings your everyday cyber operations together.
            </p>
          </div>

          <div className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((feature) => (
              <div
                key={feature.title}
                className="rounded-2xl border border-slate-200 bg-white p-6 transition hover:-translate-y-1 hover:border-blue-200 hover:shadow-lg hover:shadow-slate-900/5"
              >
                <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-50 text-lg font-bold text-[#0757B8]">
                  {feature.icon}
                </div>

                <h3 className="mt-5 text-lg font-bold text-slate-950">
                  {feature.title}
                </h3>

                <p className="mt-2 text-sm leading-6 text-slate-600">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CA COMPLIANCE */}
      <section id="compliance" className="bg-slate-50">
        <div className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
          <div className="grid gap-12 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
            <div>
              <div className="inline-flex items-center gap-2 rounded-full bg-blue-100 px-3.5 py-1.5 text-sm font-bold text-[#0757B8]">
                <ShieldCheck className="h-4 w-4" />
                CA Compliance Focus
              </div>

              <h2 className="mt-5 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">
                Be better prepared for Kenya's cyber café requirements
              </h2>

              <p className="mt-5 leading-7 text-slate-600">
                CyberSaaS is designed around the operational record-keeping
                needs of Kenyan public internet access businesses. It helps
                your cyber maintain organized customer, terminal, session,
                payment and receipt records.
              </p>

              <div className="mt-6 rounded-2xl border border-blue-100 bg-white p-5">
                <p className="text-sm font-bold text-slate-950">
                  Important
                </p>

                <p className="mt-2 text-sm leading-6 text-slate-600">
                  CyberSaaS is a compliance-support tool. Your cyber café must
                  still meet all applicable licensing, premises, operational,
                  privacy and other legal requirements.
                </p>
              </div>

              <a
                href="https://www.ca.go.ke/ca-clarifies-new-cyber-cafe-license-rules-and-says-no-browsing-history-will-be-required"
                target="_blank"
                rel="noreferrer"
                className="mt-6 inline-flex items-center text-sm font-bold text-[#0757B8] hover:underline"
              >
                Read the CA clarification
                <ChevronRight className="ml-1 h-4 w-4" />
              </a>
            </div>

            <div className="rounded-3xl border border-slate-200 bg-white p-7 shadow-xl shadow-slate-900/5 sm:p-9">
              <h3 className="text-xl font-bold text-slate-950">
                Compliance-support features
              </h3>

              <p className="mt-2 text-sm leading-6 text-slate-500">
                Designed to help cyber owners keep the operational records
                they need in an organized way.
              </p>

              <div className="mt-7 grid gap-3 sm:grid-cols-2">
                {complianceFeatures.map((item) => (
                  <div
                    key={item}
                    className="flex items-start gap-3 rounded-xl bg-slate-50 p-4"
                  >
                    <div className="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-[#0757B8] text-white">
                      <Check className="h-3 w-3" />
                    </div>

                    <span className="text-sm font-medium leading-5 text-slate-700">
                      {item}
                    </span>
                  </div>
                ))}
              </div>

              <div className="mt-7 rounded-2xl bg-[#0757B8] p-5 text-white">
                <p className="font-bold">No browsing-history storage</p>

                <p className="mt-1 text-sm leading-6 text-blue-100">
                  CyberSaaS is designed to keep the operational session
                  information needed by the system without turning your
                  business software into a browsing-history tracker.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* HOW IT WORKS */}
      <section id="how-it-works">
        <div className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
          <div className="grid gap-12 lg:grid-cols-2 lg:items-center">
            <div>
              <p className="text-sm font-bold uppercase tracking-wider text-[#0757B8]">
                How it works
              </p>

              <h2 className="mt-3 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">
                Start managing your cyber in a few steps
              </h2>

              <p className="mt-4 leading-7 text-slate-600">
                Create your Cyber Owner account, select a package, configure
                your branch and terminals, then let CyberSaaS handle the
                operational records.
              </p>
            </div>

            <div className="space-y-4">
              {[
                [
                  "01",
                  "Create your account",
                  "Register your Cyber Owner account in a few simple steps.",
                ],
                [
                  "02",
                  "Choose a package",
                  "Select the package that matches your branches and terminals.",
                ],
                [
                  "03",
                  "Set up your cyber",
                  "Add your branch, attendants, services and terminals.",
                ],
                [
                  "04",
                  "Start serving customers",
                  "Run sessions, sales and services while keeping your records organized.",
                ],
              ].map(([number, title, description]) => (
                <div
                  key={number}
                  className="flex gap-4 rounded-2xl border border-slate-200 bg-white p-5"
                >
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#0757B8] text-xs font-bold text-white">
                    {number}
                  </div>

                  <div>
                    <h3 className="font-bold text-slate-950">{title}</h3>

                    <p className="mt-1 text-sm leading-6 text-slate-600">
                      {description}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* DOWNLOAD */}
      <section id="download" className="border-y border-slate-200 bg-slate-50">
        <div className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
          <div className="overflow-hidden rounded-3xl bg-[#0757B8] shadow-2xl shadow-blue-900/15">
            <div className="grid gap-10 px-6 py-12 sm:px-10 lg:grid-cols-[1.1fr_0.9fr] lg:items-center lg:px-14 lg:py-14">
              <div className="text-white">
                <div className="inline-flex items-center gap-2 rounded-full bg-white/10 px-3.5 py-1.5 text-sm font-semibold text-white">
                  <Download className="h-4 w-4" />
                  CyberSaaS Terminal for Windows
                </div>

                <h2 className="mt-5 text-3xl font-bold tracking-tight sm:text-4xl">
                  Download CyberSaaS Terminal for your cyber computers
                </h2>

                <p className="mt-4 max-w-2xl leading-7 text-blue-100">
                  Install the CyberSaaS Terminal application on the computers
                  used by your cyber café. The application is designed for
                  cyber attendants and terminal operations.
                </p>

                <div className="mt-7 flex flex-col gap-3 sm:flex-row">
                  <a
                    href="https://github.com/leonelmukhwana/cybersaas/releases/download/v1.0.0/CyberSaaS-Terminal-Setup.exe"
                    className="inline-flex items-center justify-center rounded-xl bg-white px-6 py-3.5 text-sm font-bold text-[#0757B8] transition hover:bg-blue-50"
                  >
                    <Download className="mr-2 h-4 w-4" />
                    Download for Windows
                  </a>

                  <a
                    href="/downloads/CyberSaaS-Installation-Guide.pdf"
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center justify-center rounded-xl border border-white/30 px-6 py-3.5 text-sm font-bold text-white transition hover:bg-white/10"
                  >
                    <FileText className="mr-2 h-4 w-4" />
                    Installation Guide
                  </a>
                </div>

                <p className="mt-4 text-xs text-blue-200">
                  Windows x64 • Version 1.0.0 • Self-contained installer
                </p>

                <p className="mt-2 break-all text-[10px] leading-5 text-blue-200/80">
                  SHA-256:
                  9AB5B51EC90FBAFF74414A9ECA2766CDA1E819A5F9F850CA5EF81CEE5586D0B4
                </p>
              </div>

              <div className="rounded-2xl bg-white p-6 shadow-xl">
                <h3 className="text-lg font-bold text-slate-950">
                  Installation is simple
                </h3>

                <div className="mt-5 space-y-4">
                  {installationSteps.map((step) => (
                    <div key={step.number} className="flex gap-3">
                      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-xs font-bold text-[#0757B8]">
                        {step.number}
                      </div>

                      <div>
                        <p className="text-sm font-bold text-slate-900">
                          {step.title}
                        </p>

                        <p className="mt-1 text-xs leading-5 text-slate-500">
                          {step.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>

                <div className="mt-6 rounded-xl bg-slate-50 p-4">
                  <p className="text-xs font-bold text-slate-900">
                    Need help installing?
                  </p>

                  <a
                    href={getWhatsAppUrl()}
                    target="_blank"
                    rel="noreferrer"
                    className="mt-1 inline-flex items-center text-xs font-semibold text-[#0757B8]"
                  >
                    WhatsApp 0793 750 450
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* PRICING */}
      <section id="pricing" className="bg-white">
        <div className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
          <div className="mx-auto max-w-2xl text-center">
            <p className="text-sm font-bold uppercase tracking-wider text-[#0757B8]">
              Subscription packages
            </p>

            <h2 className="mt-3 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">
              Choose a package for your cyber
            </h2>

            <p className="mt-4 text-slate-600">
              Choose the package that matches your branches and terminals.
              Manage your subscription directly from your Cyber Owner account.
            </p>
          </div>

          <div className="mt-12 grid gap-5 lg:grid-cols-3">
            {plansLoading ? (
              <>
                {[1, 2, 3].map((item) => (
                  <div
                    key={item}
                    className="animate-pulse rounded-2xl border border-slate-200 bg-white p-7"
                  >
                    <div className="h-6 w-32 rounded bg-slate-200" />

                    <div className="mt-3 h-4 w-full rounded bg-slate-100" />
                    <div className="mt-2 h-4 w-4/5 rounded bg-slate-100" />

                    <div className="my-6 h-px bg-slate-200" />

                    <div className="space-y-3">
                      <div className="h-4 rounded bg-slate-100" />
                      <div className="h-4 rounded bg-slate-100" />
                      <div className="h-4 rounded bg-slate-100" />
                    </div>

                    <div className="mt-7 h-12 rounded-xl bg-slate-100" />
                  </div>
                ))}
              </>
            ) : plansError ? (
              <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center lg:col-span-3">
                <p className="text-sm text-slate-600">{plansError}</p>

                <button
                  type="button"
                  onClick={() => window.location.reload()}
                  className="mt-4 rounded-xl bg-[#0757B8] px-5 py-3 text-sm font-bold text-white transition hover:bg-[#064A9D]"
                >
                  Try Again
                </button>
              </div>
            ) : plans.length === 0 ? (
              <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center lg:col-span-3">
                <p className="text-sm text-slate-600">
                  No subscription packages are currently available.
                </p>
              </div>
            ) : (
              plans.map((plan, index) => {
                const isPopular =
                  !plan.is_lifetime &&
                  index ===
                    plans.filter((item) => !item.is_lifetime).length - 1;

                const price = Number(plan.monthly_price || 0);

                return (
                  <div
                    key={plan.id}
                    className={`relative rounded-2xl border bg-white p-7 ${
                      isPopular
                        ? "border-[#0757B8] shadow-xl shadow-blue-900/10"
                        : "border-slate-200"
                    }`}
                  >
                    {isPopular && (
                      <div className="absolute right-5 top-5 rounded-full bg-blue-50 px-3 py-1 text-xs font-bold text-[#0757B8]">
                        Popular
                      </div>
                    )}

                    {plan.is_lifetime && (
                      <div className="absolute right-5 top-5 rounded-full bg-blue-50 px-3 py-1 text-xs font-bold text-[#0757B8]">
                        Lifetime
                      </div>
                    )}

                    <h3 className="text-xl font-bold text-slate-950">
                      {plan.name}
                    </h3>

                    <p className="mt-2 min-h-12 text-sm leading-6 text-slate-600">
                      {plan.is_lifetime
                        ? "A one-time premium option for established cyber café businesses."
                        : `For cyber cafés with up to ${plan.included_branches} ${
                            plan.included_branches === 1
                              ? "branch"
                              : "branches"
                          }.`}
                    </p>

                    <div className="mt-5">
                      <span className="text-3xl font-extrabold text-slate-950">
                        KES {price.toLocaleString("en-KE")}
                      </span>

                      <span className="ml-1 text-sm text-slate-500">
                        {plan.is_lifetime ? "one-time" : "/ month"}
                      </span>
                    </div>

                    <div className="my-6 h-px bg-slate-200" />

                    <ul className="space-y-3">
                      <li className="flex items-center gap-2 text-sm text-slate-600">
                        <Check className="h-4 w-4 text-[#0757B8]" />
                        {plan.included_branches}{" "}
                        {plan.included_branches === 1
                          ? "branch"
                          : "branches"}
                      </li>

                      <li className="flex items-center gap-2 text-sm text-slate-600">
                        <Check className="h-4 w-4 text-[#0757B8]" />
                        {plan.included_terminals} terminals
                      </li>

                      {!plan.is_lifetime && (
                        <>
                          <li className="flex items-center gap-2 text-sm text-slate-600">
                            <Check className="h-4 w-4 text-[#0757B8]" />
                            Additional branches available
                          </li>

                          <li className="flex items-center gap-2 text-sm text-slate-600">
                            <Check className="h-4 w-4 text-[#0757B8]" />
                            Additional terminals available
                          </li>
                        </>
                      )}

                      {plan.is_lifetime && (
                        <li className="flex items-center gap-2 text-sm text-slate-600">
                          <Check className="h-4 w-4 text-[#0757B8]" />
                          One-time lifetime access
                        </li>
                      )}
                    </ul>

                    <Link
                      href={`/register/owner?plan=${encodeURIComponent(
                        plan.id
                      )}`}
                      className={`mt-7 flex w-full items-center justify-center rounded-xl px-4 py-3 text-sm font-bold transition ${
                        isPopular
                          ? "bg-[#0757B8] text-white hover:bg-[#064A9D]"
                          : "border border-slate-300 bg-white text-slate-700 hover:border-[#0757B8] hover:text-[#0757B8]"
                      }`}
                    >
                      Get Started
                    </Link>
                  </div>
                );
              })
            )}
          </div>

          <div className="mt-6 rounded-2xl border border-blue-100 bg-blue-50 p-6 text-center">
            <h3 className="font-bold text-slate-950">
              Looking for a lifetime option?
            </h3>

            <p className="mt-1 text-sm text-slate-600">
              If your account has a lifetime package available, it will appear
              automatically above.
            </p>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="px-5 py-20 lg:px-8">
        <div className="mx-auto max-w-6xl overflow-hidden rounded-3xl bg-[#0757B8] px-6 py-14 text-center text-white shadow-xl shadow-blue-900/15 sm:px-10">
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-white/10 text-2xl">
            🇰🇪
          </div>

          <h2 className="mt-5 text-3xl font-bold tracking-tight sm:text-4xl">
            Your cyber. Your business. Built for Kenya.
          </h2>

          <p className="mx-auto mt-4 max-w-2xl leading-7 text-blue-100">
            Stop depending on notebooks and scattered records. Bring your
            cyber café operations into one modern system built around the way
            Kenyan cyber cafés work.
          </p>

          <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
            <Link
              href="/register/owner"
              className="rounded-xl bg-white px-6 py-3.5 text-sm font-bold text-[#0757B8] transition hover:bg-blue-50"
            >
              Create Cyber Owner Account
            </Link>

            <a
              href={getWhatsAppUrl()}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center justify-center rounded-xl border border-white/30 px-6 py-3.5 text-sm font-bold text-white transition hover:bg-white/10"
            >
              <MessageCircle className="mr-2 h-4 w-4" />
              Talk to EPOA
            </a>
          </div>
        </div>
      </section>

      {/* CONTACT */}
      <section id="contact" className="border-t border-slate-200 bg-slate-50">
        <div className="mx-auto max-w-7xl px-5 py-16 lg:px-8">
          <div className="grid gap-8 md:grid-cols-3">
            <div>
              <div className="text-xl font-bold text-slate-950">
                Cyber<span className="text-[#0757B8]">SaaS</span>
              </div>

              <p className="mt-2 max-w-sm text-sm leading-6 text-slate-500">
                A Kenyan-built cyber café management platform by EPOA,
                designed for modern cyber café owners and attendants.
              </p>
            </div>

            <a
              href={getWhatsAppUrl()}
              target="_blank"
              rel="noreferrer"
              className="rounded-2xl border border-slate-200 bg-white p-5 transition hover:border-blue-200 hover:shadow-md"
            >
              <MessageCircle className="h-5 w-5 text-[#0757B8]" />

              <p className="mt-3 text-sm font-bold text-slate-950">
                WhatsApp / Phone
              </p>

              <p className="mt-1 text-sm text-slate-500">
                0793 750 450
              </p>
            </a>

            <a
              href="mailto:epoa2026@gmail.com"
              className="rounded-2xl border border-slate-200 bg-white p-5 transition hover:border-blue-200 hover:shadow-md"
            >
              <Mail className="h-5 w-5 text-[#0757B8]" />

              <p className="mt-3 text-sm font-bold text-slate-950">Email</p>

              <p className="mt-1 text-sm text-slate-500">
                epoa2026@gmail.com
              </p>
            </a>
          </div>
        </div>
      </section>

      {/* FOOTER */}
      <footer className="border-t border-slate-200 bg-white">
        <div className="mx-auto flex max-w-7xl flex-col gap-6 px-5 py-8 lg:flex-row lg:items-center lg:justify-between lg:px-8">
          <div>
            <div className="text-lg font-bold text-slate-950">
              Cyber<span className="text-[#0757B8]">SaaS</span>
            </div>

            <p className="mt-1 text-sm text-slate-500">
              By EPOA • Kenyan-built cyber café management software.
            </p>
          </div>

          <div className="flex flex-wrap gap-5 text-sm">
            <Link
              href="/login/owner"
              className="text-slate-500 hover:text-[#0757B8]"
            >
              Cyber Owner Login
            </Link>

            <Link
              href="/login/cyber-attendant"
              className="text-slate-500 hover:text-[#0757B8]"
            >
              Attendant Login
            </Link>

            <Link
              href="/login/saas-owner"
              className="text-slate-500 hover:text-[#0757B8]"
            >
              SaaS Owner Login
            </Link>

            <a
              href={getWhatsAppUrl()}
              target="_blank"
              rel="noreferrer"
              className="text-slate-500 hover:text-[#0757B8]"
            >
              WhatsApp
            </a>

            <a
              href="mailto:epoa2026@gmail.com"
              className="text-slate-500 hover:text-[#0757B8]"
            >
              Email
            </a>
          </div>
        </div>

        <div className="border-t border-slate-100 py-5 text-center text-xs text-slate-400">
          © {new Date().getFullYear()} EPOA. CyberSaaS. All rights reserved.
        </div>
      </footer>
    </main>
  );
}