export const CONSTANTS = {
  SEARCH: {
    MAX_LENGTH: 255,
    SEARCH_BUTTON_LABEL: "Go",
  },
};

export default function Home() {
  return (
    <main>
      <section id="hero">
        <h1>Ohana</h1>
        <h2>
          Surf the <em>living</em> web
          <noscript>
            {" "}
            <u>without Javascript</u>
          </noscript>
          .
        </h2>

        <form action="/search" method="GET">
          <input
            type="text"
            name="query"
            id="query"
            required
            maxLength={CONSTANTS.SEARCH.MAX_LENGTH}
          />
          <input type="submit" value={CONSTANTS.SEARCH.SEARCH_BUTTON_LABEL} />
        </form>
      </section>

      <section id="catalogue"></section>
    </main>
  );
}
