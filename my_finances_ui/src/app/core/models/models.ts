export interface User {
  id: string;
  name: string;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface BankAccount {
  id: string;
  user_id: string;
  name: string;
  balance: number;
  icon_url: string | null;
  created_at: string;
  updated_at: string;
}

export interface Tag {
  id: string;
  user_id: string;
  name: string;
  color: string | null;
  created_at: string;
  updated_at: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  bank_account_id: string;
  name: string;
  value: number;
  operation: 'add' | 'subtract';
  date: string;
  tags: Tag[];
  created_at: string;
  updated_at: string;
}

export interface Transfer {
  id: string;
  user_id: string;
  source_account_id: string;
  target_account_id: string;
  name: string;
  value: number;
  date: string;
  tags: Tag[];
  created_at: string;
  updated_at: string;
}

export interface Goal {
  id: string;
  user_id: string;
  source_account_id: string;
  target_account_id: string;
  name: string;
  start_date: string;
  end_date: string;
  interval_days: number;
  target_value: number;
  created_at: string;
  updated_at: string;
}

export interface Income {
  id: string;
  user_id: string;
  bank_account_id: string;
  name: string;
  value: number;
  repeatable_day: number;
  last_executed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Expense {
  id: string;
  user_id: string;
  bank_account_id: string;
  name: string;
  value: number;
  repeatable_day: number;
  last_executed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  pages: number;
}

export interface TagStat {
  tag: Tag;
  count: number;
}

export interface UpcomingItem {
  type: 'income' | 'expense' | 'goal_transfer';
  name: string;
  value: number;
  operation: 'add' | 'subtract';
  date: string;
  source: string;
}

export interface FeedEntry {
  id: string;
  name: string;
  value: number;
  operation: 'add' | 'subtract';
  date: string;
  kind: 'transaction' | 'transfer';
  tags?: Tag[];
}

export interface BankAccountSummary {
  account: BankAccount;
  latest_transactions: FeedEntry[];
  tag_stats: TagStat[];
  upcoming_expenses: UpcomingItem[];
}

export interface StatementResponse {
  account: BankAccount;
  transactions: PaginatedResponse<FeedEntry>;
  upcoming: UpcomingItem[];
}
