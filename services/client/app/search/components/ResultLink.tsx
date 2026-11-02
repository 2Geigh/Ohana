"use client";

import { FC, ReactNode } from "react";
import { logVisit } from "../actions/logVisit";

type ResultLinkProps = {
  url: string;
  className: string;
  children: ReactNode;
};

const ResultLink: FC<ResultLinkProps> = ({ url, className, children }) => {
  return (
    <a
      href={url}
      className={className}
      onClick={async (e) => {
        e.preventDefault();
        try {
          await logVisit(url);
        } catch (error) {
          console.error("Failed to log visit:", error);
        } finally {
          window.location.assign(url);
        }
        window.location.href = url;
      }}
    >
      {children}
    </a>
  );
};

export default ResultLink;
