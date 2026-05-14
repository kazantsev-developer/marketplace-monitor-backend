import express from "express";
import cors from "cors";
import helmet from "helmet";
import compression from "compression";
import morgan from "morgan";
import dotenv from "dotenv";

dotenv.config();

const app = express();
const PORT = process.env.PORT || 3000;

app.use(helmet());
app.use(compression());
app.use(morgan("combined"));
app.use(
  cors({
    origin: ["http://localhost:5173", "http://localhost:3000"],
    credentials: true,
  }),
);
app.use(express.json());

const mockWbOrders = [
  {
    date: "2025-05-10",
    g_number: "WB-12345",
    supplier_article: "ART-001",
    brand: "Nike",
    category: "Кроссовки",
    nm_id: 1001,
    tech_size: "42",
    total_price: 5490,
    warehouse_name: "Чехов",
    dest_city_name: "Москва",
    is_cancel: false,
  },
  {
    date: "2025-05-11",
    g_number: "WB-12346",
    supplier_article: "ART-002",
    brand: "Adidas",
    category: "Футболка",
    nm_id: 1002,
    tech_size: "M",
    total_price: 2990,
    warehouse_name: "Подольск",
    dest_city_name: "СПБ",
    is_cancel: true,
  },
  {
    date: "2025-05-12",
    g_number: "WB-12347",
    supplier_article: "ART-003",
    brand: "Puma",
    category: "Штаны",
    nm_id: 1003,
    tech_size: "L",
    total_price: 3990,
    warehouse_name: "Чехов",
    dest_city_name: "Москва",
    is_cancel: false,
  },
];

const mockWbRemains = [
  {
    nm_id: 1001,
    size: "42",
    barcode: "4601234567890",
    warehouse: "Чехов",
    quantity: 150,
  },
  {
    nm_id: 1002,
    size: "M",
    barcode: "4601234567891",
    warehouse: "Подольск",
    quantity: 80,
  },
  {
    nm_id: 1003,
    size: "L",
    barcode: "4601234567892",
    warehouse: "Чехов",
    quantity: 45,
  },
];

const mockWbCards = [
  {
    nm_id: 1001,
    vendor_code: "ART-001",
    brand: "Nike",
    title: "Кроссовки Nike Air",
    photos: [{ big: "https://via.placeholder.com/120" }],
    sizes: [{ tech_size: "42" }],
    updated_at: new Date().toISOString(),
  },
  {
    nm_id: 1002,
    vendor_code: "ART-002",
    brand: "Adidas",
    title: "Футболка Adidas",
    photos: [],
    sizes: [{ tech_size: "M" }],
    updated_at: new Date().toISOString(),
  },
  {
    nm_id: 1003,
    vendor_code: "ART-003",
    brand: "Puma",
    title: "Штаны Puma",
    photos: [],
    sizes: [{ tech_size: "L" }],
    updated_at: new Date().toISOString(),
  },
];

const mockOzonOrders = [
  {
    posting_number: "OZON-001",
    created_at: "2025-05-10T10:00:00Z",
    status: "delivered",
    scheme: "FBO",
    products: [{ price: 5000 }],
    financial_data: { products: [{ price: 5000 }] },
  },
  {
    posting_number: "OZON-002",
    created_at: "2025-05-11T11:00:00Z",
    status: "cancelled",
    scheme: "FBS",
    products: [{ price: 3000 }],
    financial_data: { products: [{ price: 3000 }] },
  },
];

const mockOzonRemains = [
  {
    sku: "SKU-001",
    item_code: "ART-001",
    name: "Товар Ozon 1",
    brand: "OzonBrand",
    category: "Обувь",
    fbo_visible_amount: 100,
    fbo_present_amount: 120,
  },
  {
    sku: "SKU-002",
    item_code: "ART-002",
    name: "Товар Ozon 2",
    brand: "OzonBrand",
    category: "Одежда",
    fbo_visible_amount: 50,
    fbo_present_amount: 60,
  },
];

const mockMoyskladStocks = [
  {
    product_name: "Товар МС 1",
    article: "MC-001",
    store_name: "Склад МС",
    stock: 200,
    reserve: 10,
    in_transit: 5,
    product_uuid: "uuid1",
    store_uuid: "store1",
  },
  {
    product_name: "Товар МС 2",
    article: "MC-002",
    store_name: "Склад МС",
    stock: 100,
    reserve: 5,
    in_transit: 2,
    product_uuid: "uuid2",
    store_uuid: "store1",
  },
];

const mockMoyskladStores = [{ uuid: "store1", name: "Основной склад" }];

const mockSyncLogs = [
  {
    id: 1,
    entity_type: "orders",
    status: "success",
    sync_at: new Date().toISOString(),
    records_count: 15,
    execution_time_seconds: 2.5,
  },
  {
    id: 2,
    entity_type: "remains",
    status: "error",
    sync_at: new Date(Date.now() - 86400000).toISOString(),
    records_count: 0,
    execution_time_seconds: 1.2,
    error_message: "Timeout",
  },
];

const mockDailyChart = [
  { date: "2025-05-01", wb_orders: 5, ozon_orders: 3 },
  { date: "2025-05-02", wb_orders: 7, ozon_orders: 4 },
  { date: "2025-05-03", wb_orders: 6, ozon_orders: 5 },
  { date: "2025-05-04", wb_orders: 8, ozon_orders: 6 },
  { date: "2025-05-05", wb_orders: 10, ozon_orders: 7 },
];

app.get("/api/health", (req, res) => {
  res.json({
    status: "ok",
    time: new Date().toISOString(),
    services: { api: true, mock: true },
  });
});

app.get("/api/wb/orders", (req, res) => {
  const { limit = 100, offset = 0 } = req.query;
  const data = mockWbOrders.slice(
    parseInt(offset),
    parseInt(offset) + parseInt(limit),
  );
  res.json({
    data,
    pagination: {
      total: mockWbOrders.length,
      limit: parseInt(limit),
      offset: parseInt(offset),
    },
  });
});

app.get("/api/wb/orders/stats", (req, res) => {
  const total_orders = mockWbOrders.length;
  const cancelled_orders = mockWbOrders.filter((o) => o.is_cancel).length;
  const total_revenue = mockWbOrders.reduce((sum, o) => sum + o.total_price, 0);
  const unique_products = new Set(mockWbOrders.map((o) => o.nm_id)).size;
  res.json({ total_orders, cancelled_orders, total_revenue, unique_products });
});

app.get("/api/wb/remains", (req, res) => {
  res.json(mockWbRemains);
});

app.get("/api/wb/cards", (req, res) => {
  const { limit = 50, offset = 0 } = req.query;
  const data = mockWbCards.slice(
    parseInt(offset),
    parseInt(offset) + parseInt(limit),
  );
  res.json({
    data,
    pagination: {
      total: mockWbCards.length,
      limit: parseInt(limit),
      offset: parseInt(offset),
    },
  });
});

app.get("/api/ozon/orders", (req, res) => {
  const { limit = 100, offset = 0 } = req.query;
  const data = mockOzonOrders.slice(
    parseInt(offset),
    parseInt(offset) + parseInt(limit),
  );
  res.json({
    data,
    pagination: {
      total: mockOzonOrders.length,
      limit: parseInt(limit),
      offset: parseInt(offset),
    },
  });
});

app.get("/api/ozon/remains", (req, res) => {
  res.json(mockOzonRemains);
});

app.get("/api/moysklad/stocks", (req, res) => {
  res.json(mockMoyskladStocks);
});

app.get("/api/moysklad/aggregates", (req, res) => {
  res.json(mockMoyskladStocks);
});

app.get("/api/moysklad/stores", (req, res) => {
  res.json(mockMoyskladStores);
});

app.get("/api/sync/logs", (req, res) => {
  const { limit = 50 } = req.query;
  const data = mockSyncLogs.slice(0, parseInt(limit));
  res.json(data);
});

app.get("/api/dashboard/stats", (req, res) => {
  res.json({
    wb: {
      orders: mockWbOrders.length,
      remains: mockWbRemains.length,
      cards: mockWbCards.length,
    },
    ozon: { orders: mockOzonOrders.length, remains: mockOzonRemains.length },
    moysklad: {
      total_stock: mockMoyskladStocks.reduce((s, i) => s + i.stock, 0),
    },
    sync: { last_24h: 2, success_rate: 75 },
  });
});

app.get("/api/charts/orders-daily", (req, res) => {
  res.json(mockDailyChart);
});

app.listen(PORT, () => {
  console.log(`Mock server running on port ${PORT}`);
});
