defmodule OhanaClientWeb.PageController do
  use OhanaClientWeb, :controller

  def home(conn, _params) do
    render(conn, :home)
  end

  def search(conn, _params) do
    render(conn, :search)
  end
end
