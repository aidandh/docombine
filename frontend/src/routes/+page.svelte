<script lang="ts">
    const MAX_FILES = 1000;
    const MAX_SIZE = 52428800; // 50 MB
    const SUPPORTED_TYPES = ["pdf", "doc", "docx", "ppt", "pptx"];

    let documents: File[] = [];
    let canSubmit = false;
    let error = "";

    function handleFileUpload(files: File[]) {
        error = "";
        if (files.length > MAX_FILES) {
            error = "Too many documents";
            return;
        }
        if (files.reduce((accumulator, current) => accumulator + current.size, 0) > MAX_SIZE) {
            error = "Documents exceed maximum size";
            return;
        }
        if (!files.map((file) => getFileExtension(file.name)).every((ext) => SUPPORTED_TYPES.includes(ext))) {
            error = "File type not supported"
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
        {#each documents as document}
            <li>{document.name}</li>
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
