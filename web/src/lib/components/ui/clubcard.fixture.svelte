<script lang="ts">
	import ClubCard from './ClubCard.svelte';

	let {
		club = {
			id: 'c1',
			name: 'Bouvet Padel',
			avatar_icon: '',
			avatar_color: 'forest',
			my_role: 'admin',
			roster_count: 18
		},
		onclick = () => {}
	}: { club?: App.ClubListItem; onclick?: () => void } = $props();
</script>

<ClubCard {club} {onclick} />
