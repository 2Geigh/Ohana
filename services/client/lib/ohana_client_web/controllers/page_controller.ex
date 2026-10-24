defmodule OhanaClientWeb.PageController do
  use OhanaClientWeb, :controller

  def home(conn, _params) do
    render(conn, :home)
  end
end
