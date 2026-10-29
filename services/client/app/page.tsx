export const CONSTANTS = {
  SEARCH: {
    MAX_LENGTH: 255,
    SEARCH_BUTTON_LABEL: "Go",
  },
};

export default function Home() {
  return (
    <main className="h-full flex flex-column justify-center items-center bg-white dark:bg-black">
      <section id="hero" className="flex flex-col justify-center items-center">
        <header className="text-center mb-5 select-none">
          <h1 className="text-5xl font-serif font-normal tracking-wide uppercase">
            Ohana
          </h1>
          <h2 className="font-sans text-lg leading-tight">
            Surf the <em className="underline tracking-wide">living</em> web
            <noscript>{/* <br /> */} without Javascript</noscript>.
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
            className="border border-black rounded px-1 "
          />
          <input
            type="submit"
            className="bg-blue-300 hover:bg-blue-500 hover:text-white active:bg-blue-700 ml-2 px-2 py-0.5 rounded"
            value={CONSTANTS.SEARCH.SEARCH_BUTTON_LABEL}
          />
        </form>
      </section>

      <section id="catalogue"></section>
    </main>
  );
}
