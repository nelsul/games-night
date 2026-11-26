<script lang="ts">
	import { AVATAR_SIZE, encodeCanvas, loadPhoto } from '$lib/image';
	import Icon from './Icon.svelte';

	// Shows the picked photo under a square frame: drag to move, pinch or use the slider to zoom,
	// and rotate or mirror it. What is inside the frame becomes the profile picture.
	let { file, onsave, oncancel }: { file: File; onsave: (blob: Blob) => void | Promise<void>; oncancel: () => void } =
		$props();

	const V = 280; // viewport size in CSS px
	const WORK_MAX = 2048; // the rotated/mirrored copy never needs to be bigger than this
	interface Source {
		canvas: HTMLCanvasElement;
		url: string;
		w: number;
		h: number;
	}
	let original: HTMLImageElement | null = null;
	let src = $state<Source | null>(null);
	let rotation = $state(0); // 0, 90, 180 or 270 degrees clockwise
	let flipped = $state(false); // mirrored left-right
	let error = $state('');
	let zoom = $state(1); // 1 = the photo just covers the circle
	let x = $state(0); // image top-left inside the viewport
	let y = $state(0);
	let saving = $state(false);
	let viewport = $state<HTMLDivElement | null>(null);

	// Registered by hand so preventDefault works (the page must not scroll while zooming).
	$effect(() => {
		const el = viewport;
		if (!el) return;
		el.addEventListener('wheel', wheel, { passive: false });
		return () => el.removeEventListener('wheel', wheel);
	});

	const base = $derived(src ? Math.max(V / src.w, V / src.h) : 1);
	const s = $derived(base * zoom);
	const w = $derived(src ? src.w * s : 0);
	const h = $derived(src ? src.h * s : 0);

	/** Draws the original photo rotated and mirrored into a working canvas, then recenters. */
	async function rebuild() {
		if (!original) return;
		const ow = original.naturalWidth;
		const oh = original.naturalHeight;
		const k = Math.min(1, WORK_MAX / Math.max(ow, oh));
		const dw = Math.round(ow * k);
		const dh = Math.round(oh * k);
		const sideways = rotation === 90 || rotation === 270;
		const canvas = document.createElement('canvas');
		canvas.width = sideways ? dh : dw;
		canvas.height = sideways ? dw : dh;
		const ctx = canvas.getContext('2d')!;
		ctx.imageSmoothingQuality = 'high';
		ctx.translate(canvas.width / 2, canvas.height / 2);
		// Mirror first (in screen space), then rotate: "Espelhar" always flips left-right as seen.
		if (flipped) ctx.scale(-1, 1);
		ctx.rotate((rotation * Math.PI) / 180);
		ctx.drawImage(original, -dw / 2, -dh / 2, dw, dh);
		const blob = await new Promise<Blob | null>((r) => canvas.toBlob(r, 'image/jpeg', 0.92));
		if (!blob) return;
		if (src) URL.revokeObjectURL(src.url);
		src = { canvas, url: URL.createObjectURL(blob), w: canvas.width, h: canvas.height };
		zoom = 1;
		const b = Math.max(V / canvas.width, V / canvas.height);
		x = (V - canvas.width * b) / 2;
		y = (V - canvas.height * b) / 2;
	}

	$effect(() => {
		let cancelled = false;
		loadPhoto(file)
			.then(async (i) => {
				if (cancelled) return;
				original = i;
				await rebuild();
				URL.revokeObjectURL(i.src);
			})
			.catch((e) => (error = e.message));
		return () => {
			cancelled = true;
			if (src) URL.revokeObjectURL(src.url);
		};
	});

	function rotate() {
		rotation = (rotation + 90) % 360;
		rebuild();
	}

	function flip() {
		flipped = !flipped;
		rebuild();
	}

	function clamp() {
		x = Math.min(0, Math.max(V - w, x));
		y = Math.min(0, Math.max(V - h, y));
	}

	/** Zooms keeping the viewport point (px, py) still. */
	function setZoom(next: number, px = V / 2, py = V / 2) {
		next = Math.min(4, Math.max(1, next));
		const before = s;
		const ix = (px - x) / before;
		const iy = (py - y) / before;
		zoom = next;
		const after = base * next;
		x = px - ix * after;
		y = py - iy * after;
		clamp();
	}

	// Pointer handling: one finger drags, two fingers pinch.
	const pointers = new Map<number, { x: number; y: number }>();
	let pinchStart = 0;
	let zoomStart = 1;

	function down(e: PointerEvent) {
		(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
		pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
		if (pointers.size === 2) {
			const [a, b] = [...pointers.values()];
			pinchStart = Math.hypot(a.x - b.x, a.y - b.y);
			zoomStart = zoom;
		}
	}

	function move(e: PointerEvent) {
		const prev = pointers.get(e.pointerId);
		if (!prev) return;
		const cur = { x: e.clientX, y: e.clientY };
		pointers.set(e.pointerId, cur);
		if (pointers.size === 1) {
			x += cur.x - prev.x;
			y += cur.y - prev.y;
			clamp();
		} else if (pointers.size === 2 && pinchStart > 0) {
			const [a, b] = [...pointers.values()];
			const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
			const mid = { x: (a.x + b.x) / 2 - rect.left, y: (a.y + b.y) / 2 - rect.top };
			setZoom((zoomStart * Math.hypot(a.x - b.x, a.y - b.y)) / pinchStart, mid.x, mid.y);
		}
	}

	function up(e: PointerEvent) {
		pointers.delete(e.pointerId);
		if (pointers.size < 2) pinchStart = 0;
	}

	function wheel(e: WheelEvent) {
		e.preventDefault();
		const rect = viewport!.getBoundingClientRect();
		setZoom(zoom * (e.deltaY < 0 ? 1.08 : 1 / 1.08), e.clientX - rect.left, e.clientY - rect.top);
	}

	function key(e: KeyboardEvent) {
		const step = 10;
		const moves: Record<string, [number, number]> = {
			ArrowLeft: [step, 0],
			ArrowRight: [-step, 0],
			ArrowUp: [0, step],
			ArrowDown: [0, -step]
		};
		if (moves[e.key]) {
			e.preventDefault();
			x += moves[e.key][0];
			y += moves[e.key][1];
			clamp();
		} else if (e.key === '+' || e.key === '=') setZoom(zoom * 1.1);
		else if (e.key === '-') setZoom(zoom / 1.1);
	}

	async function save() {
		if (!src) return;
		saving = true;
		try {
			const canvas = document.createElement('canvas');
			canvas.width = canvas.height = AVATAR_SIZE;
			const ctx = canvas.getContext('2d')!;
			ctx.imageSmoothingQuality = 'high';
			ctx.drawImage(src.canvas, -x / s, -y / s, V / s, V / s, 0, 0, AVATAR_SIZE, AVATAR_SIZE);
			await onsave(await encodeCanvas(canvas, 0.82));
		} catch (e) {
			error = (e as Error).message;
		} finally {
			saving = false;
		}
	}
</script>

<div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="crop-title">
	<div class="sheet">
		<h2 id="crop-title">Ajustar foto</h2>
		<p class="muted">Arraste para posicionar e use o zoom. O que estiver no quadro vira sua foto.</p>

		{#if error}
			<div class="error" role="alert">{error}</div>
		{:else}
			<div
				bind:this={viewport}
				class="viewport"
				style="width: {V}px; height: {V}px"
				role="slider"
				tabindex="0"
				aria-label="Posição da foto: use as setas para mover e + ou − para zoom"
				aria-valuemin={1}
				aria-valuemax={4}
				aria-valuenow={Math.round(zoom * 100) / 100}
				onpointerdown={down}
				onpointermove={move}
				onpointerup={up}
				onpointercancel={up}
				onkeydown={key}
			>
				{#if src}
					<img src={src.url} alt="" draggable="false" style="width: {w}px; height: {h}px; transform: translate({x}px, {y}px)" />
				{/if}
				<div class="mask" aria-hidden="true"></div>
			</div>

			<div class="tools">
				<button type="button" class="tool" onclick={rotate} disabled={!src}><Icon name="rotate" size={20} />Girar</button>
				<button type="button" class="tool" class:on={flipped} aria-pressed={flipped} onclick={flip} disabled={!src}><Icon name="flip" size={20} />Espelhar</button>
			</div>

			<label class="zoom">
				<span class="muted">Zoom</span>
				<input type="range" min="1" max="4" step="0.01" value={zoom} oninput={(e) => setZoom(Number(e.currentTarget.value))} />
			</label>
		{/if}

		<div class="actions">
			<button class="btn btn-ghost" type="button" onclick={oncancel} disabled={saving}>Cancelar</button>
			<button class="btn btn-primary" type="button" onclick={save} disabled={!src || saving}>{saving ? 'Salvando…' : 'Usar foto'}</button>
		</div>
	</div>
</div>

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 60;
		background: rgba(0, 0, 0, 0.75);
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}
	.sheet {
		width: 100%;
		max-width: 520px;
		background: var(--surface);
		border-radius: 16px 16px 0 0;
		padding: 22px 20px calc(22px + env(safe-area-inset-bottom));
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 18px;
	}
	.sheet h2 {
		margin: 0;
		font-size: 20px;
	}
	.sheet p {
		margin: 0;
		text-align: center;
	}
	.viewport {
		position: relative;
		overflow: hidden;
		border-radius: 8px;
		background: #000;
		touch-action: none;
		cursor: grab;
		user-select: none;
	}
	.viewport:active {
		cursor: grabbing;
	}
	.viewport img {
		position: absolute;
		left: 0;
		top: 0;
		transform-origin: 0 0;
		max-width: none;
		pointer-events: none;
	}
	/* Outlines the frame, with thirds to help line up a face. */
	.mask {
		position: absolute;
		inset: 0;
		border: 2px solid #fff;
		border-radius: 8px;
		pointer-events: none;
		background:
			linear-gradient(to right, transparent calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% + 0.5px), transparent calc(33.33% + 0.5px), transparent calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% + 0.5px), transparent calc(66.66% + 0.5px)),
			linear-gradient(to bottom, transparent calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% + 0.5px), transparent calc(33.33% + 0.5px), transparent calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% + 0.5px), transparent calc(66.66% + 0.5px));
	}
	.tools {
		width: 100%;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}
	.tool {
		height: 48px;
		border-radius: var(--radius);
		border: 1px solid var(--line-strong);
		background: transparent;
		color: var(--text);
		font-weight: 600;
		font-size: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
	}
	.tool.on {
		border-color: #fff;
		background: #fff;
		color: #141414;
	}
	.zoom {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.zoom input {
		flex: 1;
		accent-color: var(--brand);
		height: 44px;
	}
	.actions {
		width: 100%;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}
</style>
