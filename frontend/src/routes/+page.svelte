<!-- Draggable list adapted from https://svelte.dev/repl/e62f83d69cea4fda9e8a897f50b5a67c?version=4.2.18 -->

<script lang="ts">
    import { dev } from "$app/environment";

    const MAX_FILES = 1000;
    const MAX_SIZE = 50 * 1024 * 1024;
    const SUPPORTED_TYPES = ["pdf", "doc", "docx", "ppt", "pptx"];
    const API_URL = dev ? "http://localhost:8080/combine" : "/combine";

    let documents: File[] = [];
    let canSubmit = false;
    let error = "";

    let dragging: File | null = null;
    let draggingIndex: number | null = null;
    let hoveringIndex: number | null = null;
    $: {
        const swap = () => {
            if (
                draggingIndex === null ||
                hoveringIndex === null ||
                draggingIndex === hoveringIndex
            )
                return;
            [documents[draggingIndex], documents[hoveringIndex]] = [
                documents[hoveringIndex],
                documents[draggingIndex],
            ];
            draggingIndex = hoveringIndex;
        };
        swap();
    }

    function handleFileUpload(files: File[]) {
        error = "";
        documents = [];
        if (files.length > MAX_FILES) {
            error = "Too many documents";
            return;
        }
        if (
            files.reduce(
                (accumulator, current) => accumulator + current.size,
                0,
            ) > MAX_SIZE
        ) {
            error = "Documents exceed maximum size";
            return;
        }
        if (
            !files
                .map((file) => getFileExtension(file.name))
                .every((ext) => SUPPORTED_TYPES.includes(ext))
        ) {
            error = "File type not supported";
            return;
        }
        canSubmit = true;
        documents = files;
    }

    async function handleCombine() {
        canSubmit = false;
        const request = new FormData();
        documents.forEach((document) => request.append("documents", document));
        try {
            const res = await fetch(API_URL, {
                method: "POST",
                body: request,
            });
            if (!res.ok) {
                error = await res.text();
                canSubmit = true;
                return;
            }
            const blob = await res.blob();
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = "combined.pdf";
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        } catch (err) {
            error = err as string;
            canSubmit = true;
        }
    }

    function handleCancel() {
        documents = [];
        error = "";
    }

    function getFileExtension(filename: string) {
        let parts: string[];
        if (filename.startsWith(".")) {
            parts = filename.slice(1).split(".");
        } else {
            parts = filename.split(".");
        }
        return parts.pop() || "";
    }
</script>

<div class="header">
    <h1>Docombine</h1>
    <p>Combine multiple documents into one PDF document.</p>
    <p>
        Supported file types: {SUPPORTED_TYPES.reduce(
            (accumulator, current) => accumulator + ", " + current,
        )}
    </p>
    <p>Maximum combined size: {MAX_SIZE / 1024 / 1024} MB</p>
</div>
<div class="document-container">
    {#if documents.length === 0}
        <form class="upload-form">
            <!-- TODO: Changing this to a div and getting rid of the form might make styling easier -->
            <label
                for="document-upload"
                class="upload-label"
                on:dragover|preventDefault={(e) =>
                    e.currentTarget.classList.add("on-hover")}
                on:dragleave={(e) =>
                    e.currentTarget.classList.remove("on-hover")}
                on:drop|preventDefault={(e) => {
                    e.currentTarget.classList.remove("on-hover");
                    e.dataTransfer &&
                        handleFileUpload(Array.from(e.dataTransfer.files));
                }}
            >
                Drag files here or click to upload
            </label>
            <input
                on:change={(e) =>
                    e.currentTarget.files &&
                    handleFileUpload(Array.from(e.currentTarget.files))}
                type="file"
                id="document-upload"
                class="document-upload"
                name="documents"
                accept={SUPPORTED_TYPES.reduce(
                    (accumulator, current) => accumulator + "." + current + ",",
                    "",
                )}
                multiple
            />
        </form>
    {:else}
        <div class="buttons">
            <button on:click={handleCancel}> Cancel </button>
            <button
                on:click={handleCombine}
                disabled={!canSubmit}
                id="submit-button"
            >
                Combine Documents
            </button>
        </div>
    {/if}
    {#if error === ""}
        <ul class="document-list">
            {#each documents as document, index (document)}
                <li
                    class="document"
                    draggable="true"
                    style={dragging?.name === document.name
                        ? "opacity : 0;"
                        : ""}
                    on:dragstart={(e) => {
                        canSubmit = true;
                        dragging = document;
                        draggingIndex = index;
                    }}
                    on:dragover={() => {
                        hoveringIndex = index;
                    }}
                    on:dragend={() => {
                        dragging = null;
                        draggingIndex = null;
                        hoveringIndex = null;
                    }}
                >
                    {document.name}
                </li>
            {/each}
        </ul>
    {:else}
        <p>{error}</p>
    {/if}
</div>

<!-- to prevent style culling, probably a better way to do this -->
<div style="display: none;" class="upload-label on-hover"></div>

<style>
    /* TODO: make CSS less bad */

    * {
        font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
    }

    .header {
        display: flex;
        flex-direction: column;
        align-items: center;
        height: 5vh;
    }

    .header * {
        margin: 5px;
    }

    .document-container {
        display: flex;
        justify-content: center;
        align-items: center;
        flex-direction: column;
        height: 80vh;
        padding-top: 10vh;
    }

    .upload-form {
        padding-top: 52px;
        padding-bottom: 52px;
    }

    .document-upload {
        display: none;
    }

    .upload-label {
        box-sizing: border-box;
        border: 2px dashed black;
        padding: 50px;
        border-radius: 25px;
    }

    .upload-label.on-hover {
        border-style: solid;
    }

    .document-list {
        list-style-type: none;
        padding-left: 0px;
        height: 50%;
        width: 30%;
        display: flex;
        flex-direction: column;
        align-items: center;
        overflow: scroll;
    }

    .document {
        border: 1px solid black;
        margin: 5px;
        padding: 10px;
        width: 90%;
        text-align: center;
        border-radius: 15px;
    }

    .document:hover {
        cursor: grab;
    }

    button {
        background-color: white;
        border: 1px solid black;
        border-radius: 5px;
        padding: 5px;
    }

    button:not(:disabled):hover {
        cursor: pointer;
        background-color: lightgray;
    }
</style>
