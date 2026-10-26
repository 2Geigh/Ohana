defmodule OhanaClientWeb.PageController do
  use OhanaClientWeb, :controller

  def home(conn, _params) do
    render(conn, :home)
  end

  def search(conn, _params) do
    results = [
      %{
        link: "https://www.example.com",
        title: "Example Website!",
        description: "This is an example website. Lorem ipsum and all that lol."
      },
      %{
        link: "https://www.example.com",
        title: "Example Website!",
        description: "This is an example website. Lorem ipsum and all that lol."
      },
      %{
        link: "https://www.example.com",
        title: "Example Website!",
        description: "This is an example website. Lorem ipsum and all that lol."
      },
      %{
        link: "https://www.example.com",
        title: "Example Website!",
        description: "This is an example website. Lorem ipsum and all that lol."
      },
      %{
        link: "https://www.example.com",
        title: "Example Website!",
        description: "This is an example website. Lorem ipsum and all that lol."
      },
      %{
        link: "https://www.example.com",
        title: "Example Website!",
        description: "This is an example website. Lorem ipsum and all that lol."
      }
    ]

    render(conn, :search, results: results)
  end
end
