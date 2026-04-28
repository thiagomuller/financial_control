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
  is_system: boolean;
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
  is_repeatable: boolean;
  repeatable_day: number | null;
  projected_balance_warning?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTransactionRequest {
  name: string;
  value: number;
  operation: 'add' | 'subtract';
  bank_account_id: string;
  date: string;
  tag_ids: string[];
  is_repeatable?: boolean;
  repeatable_day?: number | null;
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
  is_repeatable: boolean;
  repeatable_day: number | null;
  projected_balance_warning?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTransferRequest {
  name: string;
  value: number;
  source_account_id: string;
  target_account_id: string;
  date: string;
  tag_ids: string[];
  is_repeatable?: boolean;
  repeatable_day?: number | null;
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
  type: 'goal_transfer' | 'repeatable_transaction' | 'repeatable_transfer' | string;
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
