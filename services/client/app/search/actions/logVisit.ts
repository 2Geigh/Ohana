"use server";

import sql from "@/lib/db/db";

export async function logVisit(url: string) {
  await sql`
        UPDATE pages
        SET visits = visits + 1
        WHERE link = ${url};
        `;

  console.log(`Updated ${url} visit number`);
}
