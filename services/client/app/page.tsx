import sql from "@/lib/db/db";
import { connection } from "next/server";
import postgres from "postgres";
import { FC, JSX } from "react";

export const CONSTANTS = {
  SEARCH: {
    MAX_LENGTH: 255,
    SEARCH_BUTTON_LABEL: "Go",
  },
};

const Hero: FC = () => {
  return (
    <section
      id="hero"
      className="flex flex-col justify-center items-center bg-red-500 min-h-1/2 w-full"
    >
      <header className="text-center mb-8 select-none">
        <h1 className="text-5xl font-serif font-normal tracking-wide uppercase mb-1 dark:text-gray-200">
          Ohana
        </h1>
        <h2 className="font-sans text-lg leading-tight dark:text-gray-200 dark:font-light">
          Surf the <em className="underline tracking-wide">living</em> web
          <noscript> without Javascript</noscript>.
        </h2>
      </header>

      <form
        action="/search"
        method="GET"
        className=" flex flex-row justify-stretch dark:border-gray-400 p-1.5 box-border "
      >
        <input
          type="text"
          name="query"
          id="query"
          required
          maxLength={CONSTANTS.SEARCH.MAX_LENGTH}
          className="border border-black dark:border-gray-400 rounded px-1 "
        />
        <input
          type="submit"
          className="bg-blue-300 dark:bg-rose-500
                        hover:bg-blue-500 dark:hover:bg-rose-600
                        hover:text-white dark:hover:text-current
                        active:bg-blue-700 dark:active:bg-rose-800
                        ml-2 px-2 py-0.5 rounded"
          value={CONSTANTS.SEARCH.SEARCH_BUTTON_LABEL}
        />
      </form>
    </section>
  );
};

const Catalogue = async () => {
  type Result = {
    title: string;
    description: string;
    link: string;
    date: string;
  };

  type SectionProps = {
    title: string;
    query: postgres.Fragment;
  };

  const Section: FC<SectionProps> = async ({ title, query }) => {
    const results: Result[] = await sql`${query}`;

    const entriesLi: JSX.Element[] = results.map((entry) => (
      <li>
        <a href="https://example.com">
          <span>{entry.title}</span>
          <span>{entry.description}</span>
          <span>{entry.link}</span>
          <span>{entry.date}</span>
        </a>
      </li>
    ));

    return (
      <section className="max-w-80 bg-blue-900">
        <h3 className="font-serif text-xl ">{title}</h3>
        <ul>{entriesLi}</ul>
      </section>
    );
  };

  return (
    <section id="catalogue" className="bg-blue-600 w-full">
      <Section
        title="Recently Added"
        query={sql`SELECT title, description, link, date_discovered
                                FROM pages
                                ORDER BY date_discovered DESC
                                LIMIT 4`}
      ></Section>
      <Section
        title="Most Popular"
        query={sql`SELECT * FROM pages LIMIT 4;`}
      ></Section>
    </section>
  );
};

export default function Home() {
  return (
    <main className="h-full flex flex-col justify-center items-center bg-white dark:bg-black">
      <Hero />
      <Catalogue />
    </main>
  );
}
