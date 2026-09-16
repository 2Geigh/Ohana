from typing import Optional
import os
import asyncpg

def connect() -> Optional[connection]:
    """ Connect to the PostgreSQL database server """
    
    conn: connection = None

    try:
        print('Connecting to the PostgreSQL database...')
        conn = .connect(
            dbname=os.getenv("DB_NAME"),
            user=os.getenv("DB_USERNAME"),
            password=os.getenv("DB_PASSWORD"),
            host=os.getenv("DB_HOST")
        )
        
        with conn.cursor() as cur:        
            print('PostgreSQL database version:')
            cur.execute('SELECT version()')
            print(cur.fetchone())
            cur.close()

        return conn    

    except (Exception, psycopg2.DatabaseError) as error:
        print(f"Connect to db failed: {error}")
        return None

def disconnect(conn):
    if conn is not None:
        conn.close()
        print('Database connection closed.')