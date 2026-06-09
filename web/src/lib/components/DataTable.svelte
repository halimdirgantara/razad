<script lang="ts">
	interface Column {
		key: string;
		label: string;
		mono?: boolean;
		width?: string;
		render?: (value: unknown, row: Record<string, unknown>) => string;
	}

	interface Props {
		columns: Column[];
		rows: Record<string, unknown>[];
		emptyMessage?: string;
		emptyAction?: string;
		onemptyclick?: () => void;
		onselect?: (row: Record<string, unknown>) => void;
		selectedId?: string;
		class?: string;
	}

	let {
		columns,
		rows,
		emptyMessage = 'No data',
		emptyAction = '',
		onemptyclick,
		onselect,
		selectedId,
		class: className = ''
	}: Props = $props();

	function cellValue(row: Record<string, unknown>, col: Column): string {
		if (col.render) return col.render(row[col.key], row);
		const v = row[col.key];
		if (v === null || v === undefined) return '—';
		return String(v);
	}
</script>

<div class="table-wrapper {className}">
	{#if rows.length === 0}
		<div class="empty-state">
			<p class="empty-message">{emptyMessage}</p>
			{#if emptyAction}
				<button class="empty-action" onclick={onemptyclick}>
					{emptyAction}
				</button>
			{/if}
		</div>
	{:else}
		<table>
			<thead>
				<tr>
					{#each columns as col}
						<th style={col.width ? `width:${col.width}` : ''}>{col.label}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each rows as row}
					<tr
						class="row"
						class:selected={selectedId && row.id === selectedId}
						onclick={onselect ? () => onselect(row) : undefined}
						role={onselect ? 'button' : undefined}
						tabindex={onselect ? '0' : undefined}
						onkeydown={onselect ? (e: KeyboardEvent) => { if (e.key === 'Enter') onselect(row); } : undefined}
					>
						{#each columns as col}
							<td class:mono={col.mono}>{cellValue(row, col)}</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>

<style>
	.table-wrapper {
		overflow-x: auto;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: var(--font-size-sm);
	}
	th {
		text-align: left;
		font-weight: var(--font-weight-semibold);
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--border);
		white-space: nowrap;
		user-select: none;
	}
	td {
		padding: var(--space-2) var(--space-3);
		border-bottom: 1px solid var(--border);
		color: var(--text);
		vertical-align: middle;
	}
	td.mono {
		font-family: var(--font-mono);
		font-size: var(--font-size-xs);
	}
	.row {
		transition: background 0.1s;
	}
	.row:hover {
		background: var(--surface-2);
	}
	.row[role="button"] {
		cursor: pointer;
	}
	.row.selected {
		background: var(--primary-bg);
		border-left: 2px solid var(--primary);
	}

	/* ---- Empty State ---- */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-8) var(--space-4);
		text-align: center;
	}
	.empty-message {
		color: var(--text-muted);
		font-size: var(--font-size-sm);
	}
	.empty-action {
		background: var(--surface-2);
		color: var(--primary);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		padding: var(--space-2) var(--space-4);
		font-size: var(--font-size-sm);
		cursor: pointer;
		transition: background 0.15s;
	}
	.empty-action:hover {
		background: var(--surface-3);
	}
</style>
