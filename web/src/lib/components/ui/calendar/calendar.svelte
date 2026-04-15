<script lang="ts">
  import { Calendar as CalendarPrimitive } from 'bits-ui';
  import type { DateValue } from '@internationalized/date';
  import { cn } from '$lib/utils';

  let {
    value = $bindable(),
    class: className = '',
    weekStartsOn = 1,
    ...restProps
  }: {
    value?: DateValue;
    class?: string;
    weekStartsOn?: 0 | 1 | 2 | 3 | 4 | 5 | 6;
    [key: string]: unknown;
  } = $props();
</script>

<CalendarPrimitive.Root
  type="single"
  bind:value
  class={cn('p-3', className)}
  weekdayFormat="short"
  fixedWeeks
  {weekStartsOn}
  {...restProps}
>
  {#snippet children({ months, weekdays })}
    <CalendarPrimitive.Header class="relative flex items-center justify-between pb-3">
      <CalendarPrimitive.PrevButton
        class="flex h-8 w-8 items-center justify-center rounded-full bg-surface-raised text-text-secondary hover:bg-border transition-colors"
      >‹</CalendarPrimitive.PrevButton>
      <CalendarPrimitive.Heading class="text-sm font-semibold" />
      <CalendarPrimitive.NextButton
        class="flex h-8 w-8 items-center justify-center rounded-full bg-surface-raised text-text-secondary hover:bg-border transition-colors"
      >›</CalendarPrimitive.NextButton>
    </CalendarPrimitive.Header>

    {#each months as month}
      <CalendarPrimitive.Grid class="w-full border-collapse">
        <CalendarPrimitive.GridHead>
          <CalendarPrimitive.GridRow class="flex">
            {#each weekdays as weekday}
              <CalendarPrimitive.HeadCell class="flex-1 pb-2 text-center text-[11px] font-semibold uppercase tracking-[0.05em] text-text-disabled">
                {weekday.slice(0, 2)}
              </CalendarPrimitive.HeadCell>
            {/each}
          </CalendarPrimitive.GridRow>
        </CalendarPrimitive.GridHead>

        <CalendarPrimitive.GridBody>
          {#each month.weeks as weekDates}
            <CalendarPrimitive.GridRow class="flex">
              {#each weekDates as date}
                <CalendarPrimitive.Cell {date} month={month.value} class="flex-1 p-0.5">
                  <CalendarPrimitive.Day
                    class="mx-auto flex h-9 w-9 cursor-pointer items-center justify-center rounded-full text-sm transition-colors
                      hover:bg-surface-raised
                      data-[selected]:bg-primary data-[selected]:font-semibold data-[selected]:text-white data-[selected]:hover:bg-primary-hover
                      data-[today]:font-semibold data-[today]:text-primary data-[today]:data-[selected]:text-white
                      data-[disabled]:pointer-events-none data-[disabled]:text-text-disabled
                      data-[outside-month]:opacity-30"
                  >
                  </CalendarPrimitive.Day>
                </CalendarPrimitive.Cell>
              {/each}
            </CalendarPrimitive.GridRow>
          {/each}
        </CalendarPrimitive.GridBody>
      </CalendarPrimitive.Grid>
    {/each}
  {/snippet}
</CalendarPrimitive.Root>
