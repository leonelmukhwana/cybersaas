

export const dynamic = 'force-dynamic';


export default function GlobalError() {
  return (
    <html lang="en">
      <body className="min-h-screen bg-slate-50">
        <main className="flex min-h-screen items-center justify-center p-4">
          <div className="w-full max-w-md rounded-xl border border-slate-200 bg-white p-6 text-center shadow-sm">
            <h2 className="text-xl font-bold text-slate-900">
              Something went wrong!
            </h2>

            <p className="mt-2 text-sm text-slate-600">
              An unexpected error occurred. Please try again.
            </p>

            <button
              type="button"
              onClick={() => window.location.reload()}
              className="mt-5 rounded-lg bg-[#0757B8] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[#064A9D]"
            >
              Try again
            </button>
          </div>
        </main>
      </body>
    </html>
  );
}