import type { Metadata } from "next";
import "./globals.css";
import { connection } from "next/server";
import { FC } from "react";

// This lets NextJS know that this page shouldn't be expected to load instantaneously
// This helps the page rendering be deterministic
export const instant = false;

export const metadata: Metadata = {
  title: "Ohana | Surf the Living Web",
  description: "A comprehensive search engine and directory of the indie web!",
};

const Footer: FC = async () => {
  await connection();

  const yearText =
    new Date().getFullYear() === 2026
      ? "2026"
      : `2026 - ${new Date().getFullYear()}`;

  return (
    <div className="flex flex-col justify-between w-full items-stretch">
      <hr className="text-gray-400" />
      <footer className="w-full py-6 px-5 inline text-center bg-white dark:bg-black text-gray-800 dark:text-gray-200 font-serif bottom-0">
        Made with{" "}
        <span aria-hidden="true">
          <span className="dark:hidden">🩵</span>
          <span className="hidden dark:inline">🩷</span>
        </span>{" "}
        in 🇨🇦 by{" "}
        <a
          className="underline text-blue-800 dark:text-emerald-400"
          href="https://nicholasgarcia.com"
        >
          Nick Garcia
        </a>
        , {yearText}.
      </footer>
    </div>
  );
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className="h-full bg-white dark:bg-black flex flex-col justify-stretch text-black dark:text-white"
    >
      <body className="flex flex-col justify-between items-stretch min-h-full border-r-orange-500">
        {children}
        <Footer />
      </body>
    </html>
  );
}
