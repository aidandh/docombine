const MAX_FILES = 1000;
const MAX_SIZE = 52428800; // 50 MB

const supportedFileTypes = [
    "pdf",
    "doc",
    "docx",
    "ppt",
    "pptx"
];

const documentList = document.getElementById("document-list");
const documentUpload = document.getElementById("document-upload");
const error = document.getElementById("error");
const submitButton = document.getElementById("submit-button");

/* 
 * Document List
 */
let draggedItem = null;
documentList.addEventListener("dragstart", (e) => {
    draggedItem = e.target;
    setTimeout(() => {
        e.target.style.display = "none";
    }, 0);
});

documentList.addEventListener("dragend", (e) => {
    setTimeout(() => {
        e.target.style.display = "";
        draggedItem = null;
    }, 0);
});

documentList.addEventListener("dragover", (e) => {
    e.preventDefault();
    const afterElement = getDragAfterElement(documentList, e.clientY);
    const currentElement = document.querySelector(".dragging");
    if (afterElement == null) {
        documentList.appendChild(draggedItem);
    }
    else {
        documentList.insertBefore(draggedItem, afterElement);
    }
});

const getDragAfterElement = (container, y) => {
    const draggableElements = [...container.querySelectorAll("li:not(.dragging)")];

    return draggableElements.reduce((closest, child) => {
        const box = child.getBoundingClientRect();
        const offset = y - box.top - box.height / 2;
        if (offset < 0 && offset > closest.offset) {
            return {
                offset: offset,
                element: child,
            };
        }
        else {
            return closest;
        }
    },
        {
            offset: Number.NEGATIVE_INFINITY,
        }
    ).element;
};

/* 
 * Document Upload
 */
documentUpload.addEventListener("change", () => {
    resetError();
    const files = documentUpload.files;
    let totalSize = 0;
    documentList.replaceChildren();
    if (files.length > MAX_FILES) {
        setError(`There are too many files (max ${MAX_FILES} files)`);
        return;
    }
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        if (!supportedFileTypes.includes(getFileExtension(file.name))) {
            setError(`${file.name} is not a supported file type`);
            documentList.replaceChildren();
            break;
        }
        totalSize += file.size;
        const newDocument = document.createElement("li");
        newDocument.draggable = true;
        newDocument.innerText = file.name;
        documentList.appendChild(newDocument);
    }
    if (totalSize > MAX_SIZE) {
        setError(`The combined size of the files is too big (max ${MAX_SIZE / 1024 / 1024} MB)`);
    }
}, false);

/*
 * Error
 */
const setError = (message) => {
    error.innerText = message;
    submitButton.disabled = true;
}

const resetError = () => {
    error.innerText = "";
    submitButton.disabled = false;
}

/*
 * Misc
 */
function getFileExtension(filename) {
    if (filename.startsWith('.')) {
        const parts = filename.slice(1).split('.');
        if (parts.length > 1) {
            return parts.pop();
        } else {
            return '';
        }
    } else {
        const parts = filename.split('.');
        if (parts.length > 1) {
            return parts.pop();
        } else {
            return '';
        }
    }
}
