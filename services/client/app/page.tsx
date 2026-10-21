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
          <input type="text" name="query" id="query" required />
          <input type="submit" value="Go" />
        </form>
      </section>

      <section id="catalogue"></section>
    </main>
  );
}
