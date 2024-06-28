<!-- Draggable list adapted from https://svelte.dev/repl/e62f83d69cea4fda9e8a897f50b5a67c?version=4.2.18 -->

<script lang="ts">
    const MAX_FILES = 1000;
    const MAX_SIZE = 52428800; // 50 MB
    const SUPPORTED_TYPES = ["pdf", "doc", "docx", "ppt", "pptx"];

    let documents: File[] = [];
    let canSubmit = false;
    let error = "";

    let dragging: File | null = null;
    let draggingIndex: number | null = null;
    let hoveringIndex: number | null = null;
    $: {
        const swap = () => {
            if (draggingIndex === null || hoveringIndex === null || draggingIndex === hoveringIndex) return;
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

<h1>Docombine</h1>
<form>
    <label for="files">Select files:</label>
    <input
        on:change={(e) =>
            e.currentTarget.files &&
            handleFileUpload(Array.from(e.currentTarget.files))}
        type="file"
        id="document-upload"
        name="documents"
        accept=".pdf,.doc,.docx,.ppt,.pptx"
        multiple
    />
    <br />
    <ul id="document-list">
        {#each documents as document, index (document)}
            <li
                draggable="true"
                style="{dragging?.name === document.name ? "opacity : 0;" : ""}"
                on:dragstart={(e) => {
                    dragging = document;
                    draggingIndex = index;
                }}
                on:dragover={(e) => {
                    hoveringIndex = index;
                }}
                on:dragend={(e) => {
                    dragging = null;
                    draggingIndex = null;
                    hoveringIndex = null;
                }}
            >
                {document.name}
            </li>
        {/each}
    </ul>
    <p id="error">{error}</p>
    <input
        disabled={!canSubmit}
        id="submit-button"
        type="submit"
        value="Combine Documents"
    />
</form>

<p>dragging?.name: {dragging?.name}</p>
<p>draggingIndex: {draggingIndex}</p>
<p>hoveringIndex: {hoveringIndex}</p>
