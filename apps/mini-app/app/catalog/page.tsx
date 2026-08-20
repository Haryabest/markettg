"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { Heading } from "@astryxdesign/core/Heading";
import { SegmentedControl, SegmentedControlItem } from "@astryxdesign/core/SegmentedControl";
import { Skeleton } from "@astryxdesign/core/Skeleton";
import { Text } from "@astryxdesign/core/Text";
import { TextInput } from "@astryxdesign/core/TextInput";
import { VStack } from "@astryxdesign/core/VStack";
import { api } from "@/lib/api";
import { ProductGrid } from "@/components/product-grid";
import { AppIcon } from "@/components/icons";
import { Crown, Gem, Gift, Search, Star } from "lucide-react";

const CATALOG_TABS = [
  { slug: "stars", label: "Stars", icon: Star },
  { slug: "premium", label: "Premium", icon: Crown },
  { slug: "gifts", label: "Подарки", icon: Gift },
  { slug: "nft", label: "NFT", icon: Gem },
] as const;

const SORT_OPTIONS = [
  { value: "popularity", label: "Популярные" },
  { value: "newest", label: "Новые" },
  { value: "price_asc", label: "Дешевле" },
  { value: "price_desc", label: "Дороже" },
] as const;

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

  const { data, isLoading, isFetching, isError } = useQuery({
    queryKey: ["products", isSearchActive ? "search" : category, debouncedQuery, activeSort],
    queryFn: () =>
      api.getProducts({
        ...(isSearchActive ? { q: debouncedQuery } : { category }),
        sort: activeSort,
        limit: "24",
      }),
    enabled: !isTyping && (!willSearch || isSearchActive),
  });

  const showSkeleton =
    (willSearch && isTyping) || (isSearchActive && (isLoading || isFetching)) || (!willSearch && isLoading);

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
            : typeof data?.total === "number"
              ? `${data.total} товаров`
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
        startIcon={<AppIcon icon={Search} size={16} />}
      />

      {!isSearchActive && !willSearch ? (
        <VStack gap={2}>
          <div className="catalog-segmented w-full">
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
                  icon={<AppIcon icon={tab.icon} size={16} />}
                />
              ))}
            </SegmentedControl>
          </div>
          <div className="catalog-segmented w-full">
            <SegmentedControl
              label="Сортировка"
              value={sort}
              onChange={(value) => updateCatalogParams(category, value as CatalogSort)}
              layout="fill"
            >
              {SORT_OPTIONS.map((option) => (
                <SegmentedControlItem key={option.value} value={option.value} label={option.label} />
              ))}
            </SegmentedControl>
          </div>
        </VStack>
      ) : null}

      <ProductGrid
        products={showSkeleton ? undefined : data?.items}
        isLoading={showSkeleton}
        isError={isError}
        onRetry={() => queryClient.invalidateQueries({ queryKey: ["products"] })}
        emptyTitle={isSearchActive ? "Ничего не найдено" : "В этой категории пусто"}
        emptyDescription={
          isSearchActive
            ? "Попробуйте «stars», «premium» или название подарка."
            : category === "gifts"
              ? "Нужен TELEGRAM_BOT_TOKEN и запущенный catalog-service. Синк идёт на сервере, не в Mini App. Проверьте .env и docker compose up catalog-service bot."
              : category === "nft"
                ? "Нужны TELEGRAM_BOT_TOKEN, TELEGRAM_API_ID и TELEGRAM_API_HASH. Запустите bot — он подтянет NFT с перепродажи Telegram."
                : "Попробуйте другую вкладку — Stars, Premium, Подарки или NFT."
        }
      />
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
