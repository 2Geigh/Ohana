import { redirect } from "next/navigation";
import { FC, JSX } from "react";
import { CONSTANTS } from "../page";
import { connection } from "next/server";
import sql from "@/lib/db/db";

type SearchResult = {
  id: number;
  score: number;
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
                  pdb.score(id) AS score,
                  title,
                  description,
                  link
                FROM pages
                WHERE
                  title       ||| ${query!}::pdb.boost(2)
                  OR description ||| ${query!}
                  OR page_text   ||| ${query!}
                ORDER BY pdb.score(id) DESC
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
          <h3 className="font-bold text-blue-800 dark:text-emerald-300 mb-2">
            {result.title}
          </h3>
          <p>{result.description}</p>
          <span className="text-gray-500 dark:text-gray-300 w-full flex flex-row justify-between items-end mt-2 select-none">
            <span className="italic select-text">{result.link}</span>
            <span className="text-gray-300 dark:text-gray-500">
              {String(Math.round(result.score * 10) / 10)}
            </span>
          </span>
        </a>
      </li>
    );
  });

  const searchDuration = performance.now() - startTime;

  console.log(results);
  const numberOfResults = results.length;

  return (
    <main className="bg-white dark:bg-black px-5 pb-5 h-auto ">
      <header className="inline-flex flex-row flex-wrap items-center justify-stretch py-5 text-center">
        <a
          href="/"
          className="flex flex-row justify-start items-end leading-none mr-4"
        >
          <h1 className="font-serif font-lighter tracking-wide leading-none uppercase text-3xl">
            Ohana
          </h1>
        </a>
        <form
          action="/search"
          method="GET"
          className=" flex flex-row justify-stretch dark:border-gray-400 p-1.5 text-left"
        >
          <input
            type="text"
            name="query"
            id="query"
            required
            maxLength={CONSTANTS.SEARCH.MAX_LENGTH}
            defaultValue={query}
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
      </header>
      <hr className="text-gray-400 mb-6" />
      <section>
        <span>
          Returned {numberOfResults} results in{" "}
          {Math.round(searchDuration) / 1000} s.
        </span>
        <ul>{ResultsLi}</ul>
      </section>
    </main>
  );
};

export default Search;
