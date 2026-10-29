import { redirect } from "next/navigation";
import { FC, JSX } from "react";
import { CONSTANTS } from "../page";
import { connection } from "next/server";
import sql from "@/lib/db/db";

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
    const MAXIMUM_DISPLAYED_DESCRIPTION_LENGTH = 50; // characters

    return (
      <li key={result.id}>
        <h3>
          <a href={result.link}>{result.title}</a>
        </h3>
        <p>
          {result.description.substring(
            0,
            MAXIMUM_DISPLAYED_DESCRIPTION_LENGTH,
          )}
          {/* replace the following conditional with tailwindcss text-overflow instead, ideally */}
          {result.description.length > MAXIMUM_DISPLAYED_DESCRIPTION_LENGTH
            ? "..."
            : ""}
        </p>
        <a href={result.link}>
          {result.link} <span>{String(result.score)}</span>
        </a>
      </li>
    );
  });

  const searchDuration = performance.now() - startTime;

  console.log(results);
  const numberOfResults = results.length;

  return (
    <main>
      <section>
        <h1>
          <a href="/">Ohana</a>
        </h1>
        <form action="/search" method="GET">
          <input
            type="text"
            id="query"
            name="query"
            maxLength={CONSTANTS.SEARCH.MAX_LENGTH}
            required
            defaultValue={query}
          />
          <input type="submit" value={CONSTANTS.SEARCH.SEARCH_BUTTON_LABEL} />
        </form>
      </section>
      <section>
        Returned {numberOfResults} results in {Math.round(searchDuration)} ms.
        <ul>{ResultsLi}</ul>
      </section>
    </main>
  );
};

export default Search;
