import { redirect } from "next/navigation";
import { FC } from "react";

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

  const numberOfResults = 0;
  const searchDuration = 0;

  return (
    <main>
      <section>
        <h1>Ohana</h1>
        <form action="/search" method="GET">
          <input
            type="text"
            minLength={0}
            maxLength={0}
            required
            value={query}
          />
          <input type="submit" value="Go" />
        </form>
      </section>
      <section>
        {numberOfResults} results in {searchDuration} ms.
      </section>
    </main>
  );
};

export default Search;
