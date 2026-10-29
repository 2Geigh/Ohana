import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Ohana | Surf the Living Web",
  description: "A comprehensive search engine and directory of the indie web!",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en" className="h-full">
      <body className="h-full bg-orange-500">{children}</body>
    </html>
  );
}
