const documentList = document.getElementById("document-list");
const documentUpload = document.getElementById("document-upload");

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
    const files = documentUpload.files;
    documentList.replaceChildren();
    for (let i = 0; i < files.length; i++) {
        const newDocument = document.createElement("li");
        newDocument.setAttribute("draggable", "true");
        newDocument.innerText = files[i].name;
        documentList.appendChild(newDocument);
    }
}, false);
