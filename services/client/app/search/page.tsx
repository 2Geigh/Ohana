import { redirect } from "next/navigation";
import { FC, JSX } from "react";
import { CONSTANTS } from "../page";
import { connection } from "next/server";
import sql from "@/lib/db/db";
import { Span } from "next/dist/trace";

// This lets NextJS know that this page shouldn't be expected to load instantaneously
// This helps the page rendering be deterministic
export const instant = false;

type SearchResult = {
  id: Number;
  score: Number;
  title: string;
  description: string;
  page_text: string;
  link: string;
};

type SearchProps = {
  searchParams: Promise<{
    query?: string;
  }>;
};

const Search: FC<SearchProps> = async ({ searchParams }) => {
  const { query } = await searchParams;

  const isQueryInvalid: boolean = query === undefined || query.trim() === "";

  if (isQueryInvalid) {
    redirect("/");
  }

  console.log(`QUERY: ${query}`);

  await connection();
  const startTime = performance.now();

  const results: SearchResult[] = await sql`
        SELECT
          id,
          pdb.score(id) AS score,
          title,
          description,
          page_text,
          link
        FROM pages
        WHERE
          title          ||| ${query!}::text
          OR description ||| ${query!}::text
          OR page_text   ||| ${query!}::text
        ORDER BY
          CASE
            WHEN title       ||| ${query!}::text THEN 1
            WHEN description ||| ${query!}::text THEN 2
            ELSE 3
          END ASC,
          score DESC
        LIMIT 10;
`;

  const ResultsLi: JSX.Element[] = results.map((result) => {
    return (
      <li className="border border-gray-400 my-6 p-3 max-w-xl w-full">
        <a
          key={result.link}
          href={result.link}
          className="h-full w-full overflow-ellipsis"
        >
          <h3 className="font-bold text-blue-800 mb-2">{result.title}</h3>
          <p>{result.description}</p>
          <span className="text-gray-500 w-full flex flex-row justify-between items-end mt-2 select-none">
            <span className="italic select-text">{result.link}</span>
            <span className="text-gray-300">{String(result.score)}</span>
          </span>
        </a>
      </li>
    );
  });

  const searchDuration = performance.now() - startTime;

  console.log(results);
  const numberOfResults = results.length;

  return (
    <main className="bg-white px-5 pb-5 h-auto min-h-full">
      <header className="flex flex-row justify-start items-center py-5">
        <a
          href="/"
          className="flex flex-row justify-start items-end leading-none mr-4"
        >
          <h1 className="font-serif font-lighter tracking-wide leading-none uppercase text-3xl">
            Ohana
          </h1>
          {/* <span className="font-sans leading-none"> */}
          {/*   Surf the <em className="underline tracking-wide">living</em> web */}
          {/*   <noscript> without Javascript</noscript>. */}
          {/* </span> */}
        </a>
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
            className="border border-black rounded px-1 "
          />
          <input
            type="submit"
            className="bg-blue-300 hover:bg-blue-500 hover:text-white active:bg-blue-700 ml-2 px-2 py-0.5 rounded"
            value={CONSTANTS.SEARCH.SEARCH_BUTTON_LABEL}
          />
        </form>
      </header>
      <hr className="text-gray-400 mb-6" />
      <section>
        <span>
          Returned {numberOfResults} results in {Math.round(searchDuration)} ms.
        </span>
        <ul>{ResultsLi}</ul>
      </section>
    </main>
  );
};

export default Search;
