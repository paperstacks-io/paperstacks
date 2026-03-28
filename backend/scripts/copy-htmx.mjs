import { copyFile, mkdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const backendDir = path.resolve(scriptDir, "..");
const sourcePath = path.join(backendDir, "node_modules", "htmx.org", "dist", "htmx.min.js");
const destinationDir = path.join(backendDir, "internal", "web", "assets");
const destinationPath = path.join(destinationDir, "htmx.min.js");

await mkdir(destinationDir, { recursive: true });
await copyFile(sourcePath, destinationPath);
