/// <reference types="@sveltejs/kit" />

declare namespace App {
  interface Locals {
    user?: Record<string, any>;
    accessToken?: string;
    authUserPromise?: Promise<Record<string, any> | undefined>;
    authRefreshPromise?: Promise<import('$lib/server/api').TokenPair>;
  }

  interface PageData {}

  interface PageState {}

  interface Platform {}
}
