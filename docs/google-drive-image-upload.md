# Google Drive Image Management — Implementation Notes

> **Current implementation status:** The API and admin product form support
> uploading one product image to personal Gmail My Drive and saving its URL in
> the existing `products.image_url` field. The multi-image product gallery and
> category-image flows described in the design below are not implemented yet;
> they remain follow-up scope. This note documents the live upload setup and
> retains the broader design requirements for that future work.

## Live upload setup

1. Enable the Google Drive API in a Google Cloud project.
2. Configure the OAuth consent screen for the Gmail account that owns the
   images, and create an OAuth client for the local authorization helper.
3. **Store the downloaded `credentials.json` outside the Git repository.**
   On macOS, a suitable location is
   `~/.config/product-catalog/credentials.json`. Create the directory with
   owner-only access and restrict the file permissions:
   ```bash
   mkdir -p ~/.config/product-catalog
   chmod 700 ~/.config/product-catalog
   # Move the downloaded file to this location, then:
   chmod 600 ~/.config/product-catalog/credentials.json
   ```
   Do not place this file under the repository (including `backend/`,
   `frontend/public/`, or `frontend/src/`), in a publicly served directory,
   or in a published build, Docker image, or GitHub repository. Do not upload
   it to a public file-sharing location.
4. The current `cmd/google-drive-auth` helper does **not** read
   `credentials.json` directly. Extract the OAuth Client ID and Client Secret
   from that file and put them in the backend's untracked `.env` as
   `GOOGLE_DRIVE_OAUTH_CLIENT_ID` and
   `GOOGLE_DRIVE_OAUTH_CLIENT_SECRET`. Keep `.env` outside version control.
5. Run `make google-drive-auth` with those environment variables loaded and
   open its printed URL in the target Gmail account. The helper requests the
   narrow `drive.file` scope, creates or reuses the `Product Catalog Project`
   folder, and prints a refresh token and root folder ID.
6. Store those values in `.env` as `GOOGLE_DRIVE_REFRESH_TOKEN` and
   `GOOGLE_DRIVE_ROOT_FOLDER_ID`, then restart the API.

For a local shell session, load `.env` before running the helper or API:

```bash
set -a
source .env
set +a
make google-drive-auth
```

The backend creates or reuses `product-images/products` beneath the configured
root folder when the first upload occurs. Uploads accept JPEG, PNG, and WebP
files up to `GOOGLE_DRIVE_MAX_UPLOAD_BYTES` (5 MiB by default). When
`GOOGLE_DRIVE_PUBLIC_IMAGES=true`, only each uploaded image is made publicly
readable. The returned stable `drive.google.com/uc` URL is stored in
`products.image_url`. Drive is not a CDN and may throttle downloads or hotlink
traffic. Replacing or clearing an image URL does not delete the old Drive file.

If the Drive OAuth settings are empty, the service starts without Drive and the
image-upload endpoint responds with `503 image storage is not configured`.
Never commit `.env`, OAuth client secrets, refresh tokens, or credential JSON.
Add a defensive ignore rule such as `credentials.json` and
`*credentials*.json` to the repository's `.gitignore`; this is a backup
safeguard, not a substitute for keeping secrets outside the repository.
If a credential file or token is accidentally committed or exposed, revoke or
rotate the affected OAuth credential/token in Google Cloud/Google Account
security settings; deleting the file in a later commit does not remove it from
Git history.

## Scope

Implement Google Drive image management from the relevant admin entity screens:

- **Product Management:** one or more images per product, exactly one main image, plus ordered additional images shown as thumbnails in the product gallery.
- **Product Category Management:** one category image per category.
- **Shared storage provider:** one Google Drive integration used by both flows, with separate folders and entity-specific authorization.

The existing runtime integration currently uploads one product image before the
product is saved and stores its URL in `products.image_url`. It does not yet
provide gallery ordering, per-product folders, category images, or Drive-file
cleanup. Preserve backward compatibility while implementing those remaining
requirements.

Target storage: the owner's **personal Gmail Google Drive (My Drive)**.

## Important authentication decision: personal Drive vs Service Account

Use **Google OAuth 2.0 on behalf of the personal Gmail owner** for My Drive uploads.

Do not assume a Service Account can own files in a personal My Drive. Service
Accounts do not have ordinary consumer Drive storage quota and cannot own
files in the same way as a user. A Service Account is a better fit for a Google
Workspace Shared Drive, not this personal-Gmail prototype.

One-time setup:
1. Create/select a Google Cloud project.
2. Enable the Google Drive API.
3. Configure the OAuth consent screen for the intended account (for a private prototype, use the appropriate testing audience and add the owner's Gmail as a test user).
4. Create an OAuth Client ID for a controlled one-time authorization flow.
5. Run a local authorization helper to obtain a refresh token for the owner's Gmail account.
6. Keep the OAuth client secret and refresh token outside Git. Never put either in React/Vite variables or browser code.

Request the narrowest practical Drive scope. Verify that the selected scope permits the app to create and manage the folders/files it owns. Do not request broad Drive access without a demonstrated need.

## Drive folder organization

Create a dedicated project root folder in My Drive:

```text
My Drive/
└── Product Catalog Project/
    └── product-images/
        ├── products/
        │   ├── {product-id-1}/
        │   │   ├── {image-uuid-1}.webp
        │   │   └── {image-uuid-2}.webp
        │   └── {product-id-2}/
        └── categories/
            ├── {category-id-1}-{image-uuid}.webp
            └── {category-id-2}-{image-uuid}.webp
```

Implementation requirements:
- Accept the project root folder ID via backend configuration.
- Create child folders idempotently (search by parent + folder name; create only if missing).
- Create a product subfolder lazily when the first image is uploaded for that product.
- Upload category images into the configured `categories/` folder.
- Never search or write outside the configured project root.
- Generate UUID-based filenames; do not trust client filenames as paths.
- Keep resolved folder IDs in memory or resolve them safely on startup.

## Existing repository integration points

- Backend dependency injection: `cmd/api/main.go`.
- Backend config: `internal/config/config.go` and `.env.example`.
- Product routes: `internal/modules/product/delivery/http/product_handler.go`.
- Category routes: `internal/modules/product/delivery/http/category_handler.go`.
- Product and Category domain contracts: `internal/modules/product/domain/product.go`.
- Admin product form/API: `frontend/src/modules/admin-products/ProductFormModal.jsx` and `frontend/src/modules/admin-products/services/adminProductApi.js`.
- Axios auth/header handling: `frontend/src/services/apiClient.js`.

Important repository-specific note: Product and Category are currently managed inside the same Go product module, but have separate route groups and permissions. Existing category permissions are `categories:view`, `categories:create`, `categories:edit`, and `categories:delete`.

## Data model

### Product gallery

Add a `product_images` table:

| Column | Type | Notes |
|---|---|---|
| id | UUID | Primary key |
| product_id | UUID | FK to products |
| drive_file_id | VARCHAR | Google Drive file ID |
| image_url | TEXT | URL used by the catalog |
| is_main | BOOLEAN | Main image flag |
| sort_order | INTEGER | Gallery order |
| created_at | TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | Last update |

Constraints and consistency:
- Add a partial unique index on `product_id` where `is_main = true`, ensuring at most one main image per product.
- Enforce that a product with images has exactly one main image through transactional application logic; a newly created product may temporarily have no images.
- When the first image is added, make it main automatically.
- Selecting another main image must update the old and new flags in one database transaction.
- Reordering must use a transaction and deterministic sort order.
- When deleting the main image, promote the first remaining image to main in the same transaction.
- Mirror the main image URL into the existing `products.image_url` field to keep old product listing/detail clients backward compatible. If no gallery image remains, set it to an empty value.
- Do not store a comma-separated list or JSON array in `products.image_url`.

### Category image

Category currently has a simple entity and separate category permissions. For the prototype, add nullable `image_url` and `drive_file_id` fields to the category record (or a one-to-one category image table if the repository's migration conventions favor it). There is one image per category, not a gallery.

Add migrations with reversible up/down scripts. Inspect the actual schema and migration conventions before writing SQL.

## API design

All routes remain under the existing JWT-protected admin route group.

### Product gallery

- `POST /api/v1/admin/products/:id/images` — multipart upload; permission `products:edit`.
- `GET /api/v1/admin/products/:id/images` — list gallery; permission `products:view`.
- `PATCH /api/v1/admin/products/:id/images/:imageId/main` — set main image; permission `products:edit`.
- `PATCH /api/v1/admin/products/:id/images/reorder` — update ordered image IDs; permission `products:edit`.
- `DELETE /api/v1/admin/products/:id/images/:imageId` — delete image relation and then remove Drive file safely; permission `products:edit`.

Product creation should remain independent of image upload: create the product first, then upload images using its ID. The React form may orchestrate this after the product is created. For edit, upload directly against the existing product ID.

### Category image

- `PUT /api/v1/admin/categories/:id/image` — multipart upload or replace; permission `categories:edit`.
- `DELETE /api/v1/admin/categories/:id/image` — remove category image relation and then remove Drive file safely; permission `categories:edit`.

Do not use a generic endpoint that allows a caller to choose an arbitrary Drive folder or entity type. Entity-specific routes make authorization and folder selection explicit.

### Upload validation

- Content type: `multipart/form-data`; field name: `file`.
- Allow JPEG, PNG, and WebP only; reject SVG and other active-content formats.
- Maximum upload size: 5 MiB initially, configurable server-side.
- Limit request body before multipart parsing.
- Inspect file content (not only extension or browser MIME type); verify extension consistency.
- Generate UUID filenames and set safe Drive metadata.
- Apply public read permission only to each image file if the public catalog needs it. Never make the entire project root public.
- Return `image_url`, `file_id`, and image metadata; never return credentials or raw provider errors.
- Use a stable image URL for the existing catalog. Document that Drive is not a CDN and may have quota/hotlink limitations.

### Safe deletion and replacement

Drive and PostgreSQL cannot participate in one atomic transaction. Avoid deleting an old Drive file before the database change succeeds.

Recommended sequence:
1. Upload the new file.
2. Commit the database relation/main-image change.
3. Delete the old Drive file after commit.
4. If step 3 fails, log a cleanup task/orphan candidate; do not roll back a valid new image.
5. For deletion, remove the database relation and record a pending Drive cleanup, then retry provider deletion safely.

For the prototype, a small cleanup-pending table or structured retry mechanism is preferable to silently leaving untracked files. Ensure a provider deletion failure does not corrupt the gallery state.

## Backend architecture

1. Create a shared storage abstraction, e.g. `ImageStorage`, with upload/delete operations and typed metadata/results.
2. Implement a Google Drive adapter using the official Google API client and OAuth2 refresh-token credentials.
3. Keep provider-specific code outside product domain/use cases.
4. Add a product image use case/repository for gallery operations and a category image use case/repository for one-to-one category image changes, reusing the same storage adapter.
5. Validate entity existence before upload.
6. Enforce existing entity-specific permissions. Product upload uses `products:edit`; category upload uses `categories:edit`.
7. Add tests for unconfigured storage, oversized upload, invalid MIME, unauthorized users, insufficient permissions, entity-not-found, Drive API errors, first-image-main behavior, main-image switching, reorder, deleting main image, category replace, and cleanup failures.
8. Log operation result and Drive file ID only. Never log credentials, tokens, file bytes, or image contents.

If Drive is not configured, the app should start normally and image endpoints should return a clear `503 image storage is not configured`.

## Frontend implementation

### Product Management form

Add an Image Gallery section:
- Multi-file selection and upload.
- Main image preview, additional thumbnail strip, and upload progress.
- Mark any thumbnail as Main Image.
- Reorder thumbnails.
- Remove a selected image with confirmation.
- First successfully uploaded image becomes main automatically.
- If creating a product, save the product first to obtain its ID, then upload selected images. Show partial failure clearly and allow retry without duplicating already uploaded files.
- When editing, load existing gallery and preserve it unless the admin explicitly changes it.
- Do not delete Drive files just because a local selection is cleared before save.

### Category Management form

Add a single image picker:
- Upload/replace category image.
- Preview current image.
- Remove image with confirmation.
- Use the category-specific API and permission.

### API client

Add explicit functions for product gallery list/upload/set-main/reorder/delete and category image upload/delete. Use `FormData` and let Axios/browser set the multipart boundary. Preserve existing product/category create and update JSON contracts except for the additive category image fields if selected.

Add frontend tests for gallery interactions, upload progress/errors, main selection, reorder, deletion confirmation, product create-then-upload behavior, and category image replacement.

## Environment variables (proposed)

```dotenv
# Leave empty to disable Google Drive image management.
GOOGLE_DRIVE_OAUTH_CLIENT_ID=
GOOGLE_DRIVE_OAUTH_CLIENT_SECRET=
GOOGLE_DRIVE_REFRESH_TOKEN=
GOOGLE_DRIVE_ROOT_FOLDER_ID=
GOOGLE_DRIVE_PUBLIC_IMAGES=true
GOOGLE_DRIVE_MAX_UPLOAD_BYTES=5242880
```

Secrets must remain server-side. Add empty keys to `.env.example` and document setup. Never commit a real `.env`, OAuth client secret, refresh token, or credential JSON.

## Acceptance criteria

- Product can have multiple images and exactly one main image whenever its gallery is non-empty.
- Product main image is reflected in the existing `products.image_url` field for backward compatibility.
- Public product listing uses the main image; product detail exposes the full ordered gallery.
- Category supports one image.
- Product and Category images are uploaded into their respective configured Drive folders.
- Product image routes require `products:edit`; category image routes require `categories:edit`.
- Existing product/category create/update behavior remains backward compatible.
- Drive configuration can be absent without preventing API startup.
- Provider failures do not leave the database pointing at a deleted image.
- `go test ./...`, frontend tests, and `npm run build` pass.

## Suggested VS Code Copilot Agent prompt

> Implement the Google Drive image management described in `docs/google-drive-image-upload.md`. First inspect the existing architecture, migrations, permissions, and tests. Use OAuth 2.0 for the personal Gmail owner's My Drive, not a Service Account. Build one shared Google Drive storage adapter behind an interface. Product Management must support a gallery with multiple images, exactly one main image, ordered thumbnails, set-main, reorder, and delete. Keep the existing `products.image_url` field synchronized with the main image for backward compatibility, and expose the ordered gallery on product detail. Category Management must support one image. Use entity-specific JWT-protected routes and existing permissions: `products:edit` and `categories:edit`. Validate actual file content, allow JPEG/PNG/WebP only, enforce a configurable 5 MiB limit, and upload only inside the configured project folders. Add reversible database migrations, safe Drive cleanup behavior, frontend controls in the respective Product and Category forms, and comprehensive tests. Never commit credentials. Run Go tests, frontend tests, and frontend build; report any checks that cannot be run rather than claiming they passed. Do not merge the PR automatically.
