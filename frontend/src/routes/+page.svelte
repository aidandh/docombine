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
            const res = await fetch("http://localhost:8080/combine", {
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
        accept={SUPPORTED_TYPES.reduce(
            (accumulator, current) => accumulator + "." + current + ",",
            "",
        )}
        multiple
    />
    <br />
    <ul id="document-list">
        {#each documents as document, index (document)}
            <li
                draggable="true"
                style={dragging?.name === document.name ? "opacity : 0;" : ""}
                on:dragstart={(e) => {
                    canSubmit = true;
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
        on:click|preventDefault={handleCombine}
        disabled={!canSubmit}
        id="submit-button"
        type="submit"
        value="Combine Documents"
    />
</form>
