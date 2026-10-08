# Arcmira YouTube transcript source shortlist for n8n

[Arcmira](https://arcmira.com) · [API docs](https://arcmira.com/docs) · [Plans and usage](https://arcmira.com/docs/usage-and-billing)

Find up to three spoken passages about a topic in indexed YouTube videos. Keep each passage's source link, timestamp, publication date and coverage notes for editorial review. This workflow uses built-in n8n nodes. No community-node installation or model call is required.

## Run it

1. Import [arcmira-source-shortlist.n8n.json](arcmira-source-shortlist.n8n.json) into n8n using **Import from File**. The workflow is inactive and starts manually. It was executed with n8n 2.41.3; the browser-editor import has not been tested.
2. Create an **HTTP Bearer Auth** credential using your own [Arcmira API key](https://arcmira.com/docs/authentication). Enter the key as the token, without an extra `Bearer ` prefix. Select this credential in **Search Arcmira**. Keep it out of workflow fields, code and exports.
3. In **Search input**, set `topic`, `after` and `before`. The example searches `creativity` between August 1 and September 1, 2026. `after` is inclusive; `before` is exclusive. Dates are UTC. Choose dates available under your plan.
4. Execute once and inspect **Build source shortlist**. Keep `requested_window`, `applied_window`, `partial`, `search_index` and `note` with the candidates. Errors stop the workflow. Resolve a refused request explicitly; it does not retry or change the query for you.

The HTTP node makes one `GET https://api.arcmira.com/v1/search` with a fixed `limit=3`. There is no pagination, retry, Premium transcript read, monitor write or media download. Repeated manual executions are separate requests.

The current API makes the first five search passages in each request free. Free accounts include 1,000 monthly credits with no card, withhold the newest 30 days of media, and expose only the first page of lists. Limits and service terms apply; see the usage link above. The client workflow is open source, but it requires the hosted service and your account.

## Read the output

The workflow returns one JSON item with coverage metadata and up to three `candidates`. Each keeps the video ID, title, channel, publication date, passage text, passage start, API watch link, original YouTube link, transcript source and any available speaker labels. It preserves relative API watch links in `source_watch_url` and also supplies an absolute `arcmira_url`.

A passage timestamp is a source-navigation hint, not a verified edit boundary. Cut start/end remain null, playback and speaker identity remain unverified, and reuse permission is not established. Before publishing or editing, watch the source, check the quote and attribution, choose exact cut boundaries and establish any required rights. Speaker labels may be absent. Empty results do not prove a topic was never discussed. Search covers indexed material, not every YouTube video.

## Verification

On October 7, 2026, the actual workflow ran once with a Free account in isolated n8n 2.41.3. It returned three passages, retained the applied date window and source fields, and used zero measured credits with on-demand disabled. The response was not partial, but its search index reported `catching_up`. The live response, transcript excerpts, credential and raw execution state are not in this package.

`verification.json` records bounded results and the container image digest. Packaging removes only the local workflow identifier and empty pin/tag metadata; execution code and wiring are unchanged. No fresh live run occurred during packaging. Browser-editor validation, broader queries and WorkflowHub registration remain unverified. Results and the index can change on later executions.

Run `node validate-packet.mjs` to check the distributed workflow's digest, credential placeholders, fixed endpoint/limit and absence of execution state. This is a local packaging check, not a live API test.

## License

Copyright 2026 Arcmira. The workflow and accompanying instructions are licensed under Apache-2.0, matching the public integrations repository. See `LICENSE`. That license does not grant rights to third-party video or transcript content or replace Arcmira's service terms.
