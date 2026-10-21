import Image from "next/image";

export default function Home() {
  return (
    <main>
      <section id="hero">
        <h1>Ohana</h1>
        <h2>
          Surf the <em>living</em> web
          <noscript>
            {" "}
            <em>without</em> JavaScript
          </noscript>
        </h2>

        <form action="/search" method="get">
          <input type="text" name="q" id="q" required />
          <input type="submit" value="Go" />
        </form>
      </section>

      <section id="catalogue"></section>
    </main>
  );
}
