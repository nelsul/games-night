// Profile photos are cropped to a centered square and shrunk in the browser before
// upload, so phones on mobile data send ~100 KB instead of a 5 MB camera picture.

const SIDE = 512;

export async function squareAvatar(file: File): Promise<Blob> {
	const bitmap = await loadImage(file);
	const side = Math.min(bitmap.width, bitmap.height);
	const sx = (bitmap.width - side) / 2;
	const sy = (bitmap.height - side) / 2;
	const out = Math.min(SIDE, side);
	const canvas = document.createElement('canvas');
	canvas.width = out;
	canvas.height = out;
	const ctx = canvas.getContext('2d');
	if (!ctx) throw new Error('canvas');
	ctx.drawImage(bitmap, sx, sy, side, side, 0, 0, out, out);
	return await new Promise<Blob>((resolve, reject) =>
		canvas.toBlob((b) => (b ? resolve(b) : reject(new Error('toBlob'))), 'image/jpeg', 0.85)
	);
}

async function loadImage(file: File): Promise<ImageBitmap | HTMLImageElement> {
	if ('createImageBitmap' in window) {
		try {
			return await createImageBitmap(file, { imageOrientation: 'from-image' });
		} catch {
			/* fall back to <img>, which Safari decodes HEIC with */
		}
	}
	const url = URL.createObjectURL(file);
	try {
		const img = new Image();
		img.src = url;
		await img.decode();
		return img;
	} finally {
		URL.revokeObjectURL(url);
	}
}
