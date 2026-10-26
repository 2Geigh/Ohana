import { redirect } from "next/navigation";
import { FC } from "react";
import { CONSTANTS } from "../page";
import { connection } from "next/server";
import sql from "@/lib/db/db";

// This lets NextJS know that this page shouldn't be expected to load instantaneously
// This helps the page rendering be deterministic
export const instant = false;

async function sleep(milliseconds: number) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

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

  const results = await sql`
        SELECT
      pdb.score(id),
      title,
      description,
      link
    FROM pages
    WHERE description ||| ${query!}::text
    ORDER BY score DESC
    LIMIT 10;
`;

  const searchDuration = performance.now() - startTime;

  console.log(results);
  const numberOfResults = results.length;

  return (
    <main>
      <section>
        <h1>Ohana</h1>
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
      </section>
    </main>
  );
};

export default Search;
