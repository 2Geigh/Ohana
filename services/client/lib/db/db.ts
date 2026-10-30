import "dotenv/config";
import postgres from "postgres";

// const DB_CONNECTION_STRING = `postgres://${process.env.DB_USERNAME}:${process.env.DB_PASSWORD}@${process.env.DB_HOST}:${process.env.DB_CONTAINER_PORT}/${process.env.DB_NAME}`;

const { DB_HOST, DB_NAME, DB_USERNAME, DB_PASSWORD } = process.env;

const sql = postgres({
  database: DB_NAME,
  host: DB_HOST,
  password: DB_PASSWORD,
  user: DB_USERNAME,
});

export default sql;
