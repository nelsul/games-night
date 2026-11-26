import { api, type Me } from './api';

export const session = $state<{ me: Me | null; loaded: boolean }>({ me: null, loaded: false });

export async function loadMe(): Promise<Me | null> {
	try {
		session.me = await api.get<Me>('/me');
	} catch {
		session.me = null;
	}
	session.loaded = true;
	return session.me;
}
