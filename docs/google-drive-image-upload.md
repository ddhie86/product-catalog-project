# Google Drive Product Image Upload — Implementation Plan

## Status and scope

This document is the implementation specification for the prototype's product-image upload flow. It is intentionally a plan rather than a claim that the integration is already working. The repository's current product model already has `image_url`; the admin form currently accepts a manually entered URL. Preserve that API/database contract and replace the manual URL workflow with an upload + preview workflow.

Target storage for this prototype: the owner's **personal Gmail Google Drive (My Drive)**.

## Important authentication decision: personal Drive vs Service Account

Use **Google OAuth 2.0 on behalf of the personal Gmail owner** for My Drive uploads.

Do not assume a Service Account can own files in a personal My Drive. Service Accounts do not have ordinary consumer Drive storage quota and cannot own files in the same way as a user. A Service Account is a better fit for a Google Workspace Shared Drive, not this personal-Gmail prototype.

One-time setup:
1. Create/select a Google Cloud project.
2. Enable the Google Drive API.
3. Configure the OAuth consent screen for the intended account (for a private prototype, use the appropriate testing audience and add the owner's Gmail as a test user).
4. Create an OAuth Client ID for a Desktop app or another controlled one-time authorization flow.
5. Run a small local authorization helper to obtain a refresh token for the owner's Gmail account.
6. Keep the OAuth client secret and refresh token outside Git. Never put either in React/Vite variables or browser code.

Request the narrowest practical Drive scope. If using `drive.file`, verify that the app can create and manage the files/folders it creates. The implementation should not request broad Drive access without a demonstrated need.

## Drive folder organization

Create a dedicated project root folder in My Drive, for example:

```text
My Drive/
└── Product Catalog Project/
    └── product-images/
        ├── products/
        ├── categories/
        └── thumbnails/
```

The implementation should:
- Accept the root folder ID via backend configuration.
- Create the child folders idempotently (search by parent + folder name; create only if missing).
- Save the resolved child folder IDs in memory for the process lifetime, or resolve them safely on startup.
- Upload product images into `products/`.
- Never search or write outside the configured project root.
- Avoid hard-coding a user's Drive folder ID in source code.

## Existing repository integration points

- Backend entry point and dependency injection: `cmd/api/main.go`.
- Backend environment config: `internal/config/config.go` and `.env.example`.
- Product admin routes: `internal/modules/product/delivery/http/product_handler.go`.
- Product JSON contract already contains `image_url`: `internal/modules/product/domain/product.go`.
- Admin React form: `frontend/src/modules/admin-products/ProductFormModal.jsx`.
- Admin API client: `frontend/src/modules/admin-products/services/adminProductApi.js`.
- Axios auth/header handling: `frontend/src/services/apiClient.js`.

Keep the existing product create/update endpoints and their JSON payloads unchanged. The image upload is a separate endpoint; its response supplies an `image_url` which is then included in the normal product save request.

## Proposed API contract

### Upload

`POST /api/v1/admin/product-images`

- Requires the existing admin JWT middleware.
- Authorization must allow a user with either `products:create` OR `products:edit`, because the same form is used for both creating and editing products. Do not weaken authorization to any authenticated admin.
- Content type: `multipart/form-data`.
- Field name: `file`.
- Initial prototype limits: JPEG, PNG, and WebP only; maximum 5 MiB; reject SVG and other active-content formats.
- Validate by inspecting file content (not only the submitted MIME type or extension).
- Generate a safe unique Drive filename; do not trust the client filename as a path.
- Upload into the configured `products/` folder.
- Apply public read access only if the catalog is intended to be public. Do not make the project root or all Drive contents public.
- Return a small JSON object, e.g. `{ "image_url": "...", "file_id": "...", "name": "..." }`.
- Use a stable image URL that can be used by the existing `image_url` field. For the prototype, the Drive file must be readable by the public catalog if images are public. Document that Drive's `uc?export=view` URL is not a CDN and may have quota/hotlink limitations.
- On failure, return a sanitized API error; never return OAuth tokens, client secrets, or raw provider response bodies.

### Delete/replace behavior

Do not delete the old Drive image automatically when a product image is replaced in this first iteration. Product updates may fail after upload, and immediate deletion could break the current product image. Add orphan cleanup as a separate future task after a reliable product-to-file lifecycle is designed.

## Backend implementation tasks

1. Add a small storage abstraction in the product infrastructure/delivery boundary (or a dedicated platform package), e.g. `ImageStorage` with `Upload(ctx, reader, metadata) -> result`.
2. Implement a Google Drive adapter using the official Google API client and OAuth2 refresh-token credentials.
3. Add configuration fields for OAuth client ID, OAuth client secret, refresh token (or a protected token-file path), and project root folder ID. Validate all-or-none configuration; when not configured, the app should start normally and the upload endpoint should return a clear `503 image storage is not configured`.
4. Add a maximum request body size before multipart parsing, then validate file size, detected MIME type, and extension consistency.
5. Register the upload handler under the existing JWT-protected admin group. Add explicit OR permission handling for `products:create` and `products:edit` (or define a narrowly scoped permission and update default role seeds consistently).
6. Keep Google Drive calls out of the product domain/use case. The storage adapter is infrastructure, not product business logic.
7. Add tests for: unconfigured storage, oversized upload, invalid MIME, unauthorized admin, insufficient permissions, Drive API error mapping, and successful upload using a fake storage implementation.
8. Add structured logs for operation outcome and file ID only. Do not log file bytes, OAuth credentials, refresh tokens, or image contents.

## Frontend implementation tasks

1. Add `uploadProductImage(file)` to `adminProductApi.js`, sending `FormData` to `/admin/product-images`. Do not manually set the multipart boundary; let Axios/browser set it.
2. Update `ProductFormModal.jsx`:
   - Replace the plain URL input with a file input and an upload button.
   - Show upload progress/loading and a preview.
   - Store the returned `image_url` in the existing form state.
   - Keep the existing `image_url` value when editing a product that already has an image.
   - Allow removing the selected image from the form without deleting the Drive file.
   - Display clear validation/API errors.
   - Disable duplicate uploads while one is in progress.
3. Preserve the existing create/update payload shape and query invalidation.
4. Add frontend tests for valid selection, upload success, upload error, preview, and create/edit form behavior.

## Environment variables (names are proposed)

```dotenv
# Leave empty to disable image uploads locally.
GOOGLE_DRIVE_OAUTH_CLIENT_ID=
GOOGLE_DRIVE_OAUTH_CLIENT_SECRET=
GOOGLE_DRIVE_REFRESH_TOKEN=
GOOGLE_DRIVE_ROOT_FOLDER_ID=
GOOGLE_DRIVE_PUBLIC_IMAGES=true
GOOGLE_DRIVE_MAX_UPLOAD_BYTES=5242880
```

Use server-side secrets only. Add these keys to `.env.example` with empty values and document setup. Never commit a real `.env`, OAuth client secret, refresh token, or downloaded credential JSON.

## Acceptance criteria

- Existing catalog and product create/update API contracts remain backward compatible.
- A product admin with `products:create` or `products:edit` can upload a supported image.
- An admin without either permission cannot upload.
- Uploaded files appear only under the configured Product Catalog project folder.
- Public catalog product cards and detail pages can render the returned `image_url`.
- No database migration is required.
- Upload is disabled gracefully when Drive configuration is absent.
- `go test ./...`, frontend tests, and `npm run build` pass.
- Setup instructions explain Google Cloud OAuth consent and one-time refresh-token authorization for a personal Gmail account.

## Suggested VS Code Copilot Agent prompt

> Implement the Google Drive product-image upload described in `docs/google-drive-image-upload.md`. First inspect the existing architecture and tests. Use OAuth 2.0 for the personal Gmail owner's My Drive (not a Service Account). Keep Google Drive integration in an infrastructure adapter behind an interface. Add a JWT-protected multipart upload endpoint with an explicit OR check for `products:create` or `products:edit`. Validate actual file content, allow only JPEG/PNG/WebP, enforce a 5 MiB limit, and upload only into the configured project subfolder. Keep the existing product `image_url` contract and product create/update endpoints unchanged. Update the React admin form to upload, preview, and save the returned URL. Add tests and setup documentation. Never commit credentials. Run Go tests, frontend tests, and frontend build; report any checks that cannot be run rather than claiming they passed.
