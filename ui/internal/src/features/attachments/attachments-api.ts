export async function uploadImageAttachment(file: File, fallbackName: string) {
    const formData = new FormData();
    formData.append("file", file, file.name || fallbackName);

    const response = await fetch("/api/v1/attachments/images", {
        method: "POST",
        body: formData,
    });

    if (!response.ok) {
        throw new Error(`upload failed with status ${response.status}`);
    }

    return response.json() as Promise<string>;
}
