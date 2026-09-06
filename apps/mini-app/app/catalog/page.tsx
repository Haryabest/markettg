"use client";
import { Button, Heading, HStack, SegmentedControl, SegmentedControlItem, Skeleton, Text, TextInput, VStack } from "@/components/ui";

import { useQuery, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { ProductGrid } from "@/components/product-grid";
import { SortMenu } from "@/components/sort-menu";
import { ActionIcon } from "@/components/action-icon";

const CATALOG_TABS = [
  { slug: "stars", label: "Stars", icon: "star" },
  { slug: "premium", label: "Premium", icon: "user-star" },
  { slug: "gifts", label: "Подарки", icon: "gift" },
] as const;

const SORT_OPTIONS = [
  { value: "popularity", label: "Популярные" },
  { value: "newest", label: "Новые" },
  { value: "price_asc", label: "Дешевле" },
  { value: "price_desc", label: "Дороже" },
] as const;

const PAGE_SIZE = "100";

type CatalogTabSlug = (typeof CATALOG_TABS)[number]["slug"];
type CatalogSort = (typeof SORT_OPTIONS)[number]["value"];

function isCatalogTab(slug: string): slug is CatalogTabSlug {
  return CATALOG_TABS.some((tab) => tab.slug === slug);
}

function isCatalogSort(value: string): value is CatalogSort {
  return SORT_OPTIONS.some((option) => option.value === value);
}

function CatalogContent() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const params = useSearchParams();
  const categoryParam = params.get("category") || "stars";
  const category = isCatalogTab(categoryParam) ? categoryParam : "stars";
  const sortParam = params.get("sort") || "popularity";
  const sort: CatalogSort = isCatalogSort(sortParam) ? sortParam : "popularity";

  useEffect(() => {
    if (params.get("category") === "nft") {
      router.replace(`/catalog?category=gifts&sort=${sortParam}`);
    }
  }, [params, router, sortParam]);

  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [isTyping, setIsTyping] = useState(false);

  const trimmedQuery = query.trim();
  const willSearch = trimmedQuery.length >= 2;
  const isSearchActive = debouncedQuery.length >= 2;

  useEffect(() => {
    if (!trimmedQuery) {
      setDebouncedQuery("");
      setIsTyping(false);
      return;
    }
    if (trimmedQuery.length < 2) {
      setDebouncedQuery("");
      setIsTyping(false);
      return;
    }
    setIsTyping(true);
    const timer = window.setTimeout(() => {
      setDebouncedQuery(trimmedQuery);
      setIsTyping(false);
    }, 500);
    return () => window.clearTimeout(timer);
  }, [trimmedQuery]);

  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: () => api.getCategories(),
  });

  const activeSort = isSearchActive ? "relevance" : sort;

  const {
    data,
    isLoading,
    isFetching,
    isError,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["products", isSearchActive ? "search" : category, debouncedQuery, activeSort],
    initialPageParam: 1,
    queryFn: ({ pageParam }) =>
      api.getProducts({
        ...(isSearchActive ? { q: debouncedQuery } : { category }),
        sort: activeSort,
        limit: PAGE_SIZE,
        page: String(pageParam),
      }),
    getNextPageParam: (lastPage) => {
      const totalPages = Math.ceil(lastPage.total / lastPage.limit);
      return lastPage.page < totalPages ? lastPage.page + 1 : undefined;
    },
    enabled: !isTyping && (!willSearch || isSearchActive),
  });

  const products = data?.pages.flatMap((page) => page.items) ?? [];
  const total = data?.pages[0]?.total;

  const showSkeleton =
    (willSearch && isTyping) ||
    (isSearchActive && (isLoading || (isFetching && !products.length))) ||
    (!willSearch && isLoading);

  const updateCatalogParams = (nextCategory: CatalogTabSlug, nextSort: CatalogSort) => {
    router.push(`/catalog?category=${nextCategory}&sort=${nextSort}`);
  };

  const activeCategory = categories?.items?.find((item) => item.slug === category);
  const activeTab = CATALOG_TABS.find((tab) => tab.slug === category);

  return (
    <VStack gap={4}>
      <VStack gap={1}>
        <Heading level={1}>
          {isSearchActive ? "Поиск" : activeCategory?.name || activeTab?.label || "Каталог"}
        </Heading>
        <Text type="supporting" color="secondary" display="block">
          {isSearchActive
            ? `Результаты по запросу «${debouncedQuery}»`
            : willSearch && isTyping
              ? "Ищем после того, как вы закончите ввод…"
              : activeCategory?.description ||
                "Stars, Premium и Telegram Gifts. Доставка сразу после оплаты."}
        </Text>
        <Text type="supporting" color="secondary" display="block">
          {showSkeleton
            ? "Загрузка…"
            : typeof total === "number"
              ? `${total} товаров${products.length < total ? ` · показано ${products.length}` : ""}`
              : "Каталог"}
        </Text>
      </VStack>

      <TextInput
        label="Поиск"
        isLabelHidden
        value={query}
        onChange={setQuery}
        placeholder="Premium, Stars, подарок..."
        hasClear
        width="100%"
        startIcon={<ActionIcon name="search" size={16} />}
      />

      {!isSearchActive && !willSearch ? (
        <VStack gap={2}>
          <HStack justify="between" align="center" gap={2}>
            <div className="catalog-segmented min-w-0 flex-1">
              <SegmentedControl
                label="Категория"
                value={category}
                onChange={(value) => updateCatalogParams(value as CatalogTabSlug, sort)}
                layout="fill"
              >
                {CATALOG_TABS.map((tab) => (
                  <SegmentedControlItem
                    key={tab.slug}
                    value={tab.slug}
                    label={tab.label}
                    icon={<ActionIcon name={tab.icon} size={16} />}
                    activeValue={category}
                    onSelect={(value) => updateCatalogParams(value as CatalogTabSlug, sort)}
                  />
                ))}
              </SegmentedControl>
            </div>
            <SortMenu
              value={sort}
              options={SORT_OPTIONS}
              onChange={(value) => updateCatalogParams(category, value as CatalogSort)}
            />
          </HStack>
        </VStack>
      ) : null}

      <ProductGrid
        products={showSkeleton ? undefined : products}
        isLoading={showSkeleton}
        isError={isError}
        onRetry={() => queryClient.invalidateQueries({ queryKey: ["products"] })}
        emptyTitle={isSearchActive ? "Ничего не найдено" : "В этой категории пусто"}
        emptyDescription={
          isSearchActive
            ? "Попробуйте «stars», «premium» или название подарка."
            : category === "gifts"
              ? "Запустите bot — он подтянет весь каталог подарков Telegram (~150 шт.)."
              : "Попробуйте другую вкладку — Stars, Premium или Подарки."
        }
      />

      {!showSkeleton && hasNextPage ? (
        <Button
          label={isFetchingNextPage ? "Загрузка…" : "Показать ещё"}
          variant="secondary"
          width="100%"
          clickAction={() => {
            void fetchNextPage();
          }}
        />
      ) : null}
    </VStack>
  );
}

export default function CatalogPage() {
  return (
    <Suspense fallback={<Skeleton height={420} />}>
      <CatalogContent />
    </Suspense>
  );
}
