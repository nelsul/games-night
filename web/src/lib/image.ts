// Profile photos are cropped in the browser (AvatarCropper) and shrunk before upload,
// so phones on mobile data send ~100 KB instead of a 5 MB camera picture.

export const AVATAR_SIZE = 512;

/** Decodes a picked file into an <img>. Safari also decodes iPhone HEIC photos here. */
export async function loadPhoto(file: File): Promise<HTMLImageElement> {
	const url = URL.createObjectURL(file);
	const img = new Image();
	img.src = url;
	try {
		await img.decode();
	} catch {
		URL.revokeObjectURL(url);
		throw new Error('Não deu para abrir essa foto. Tente uma em JPG ou PNG.');
	}
	return img; // the caller revokes img.src when done
}

export function encodeCanvas(canvas: HTMLCanvasElement, quality = 0.85): Promise<Blob> {
	return new Promise((resolve, reject) =>
		canvas.toBlob((b) => (b ? resolve(b) : reject(new Error('Não deu para salvar a foto.'))), 'image/jpeg', quality)
	);
}
