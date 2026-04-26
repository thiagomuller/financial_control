import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
  {
    path: 'login',
    loadComponent: () =>
      import('./features/auth/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'register',
    loadComponent: () =>
      import('./features/auth/register/register.component').then((m) => m.RegisterComponent),
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/dashboard/dashboard.component').then((m) => m.DashboardComponent),
  },
  {
    path: 'bank-accounts',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/bank-accounts/bank-accounts.component').then(
        (m) => m.BankAccountsComponent,
      ),
  },
  {
    path: 'bank-accounts/:id',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/bank-account-detail/bank-account-detail.component').then(
        (m) => m.BankAccountDetailComponent,
      ),
  },
  {
    path: 'tags',
    canActivate: [authGuard],
    loadComponent: () => import('./features/tags/tags.component').then((m) => m.TagsComponent),
  },
  {
    path: 'transactions',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/transactions/transactions.component').then((m) => m.TransactionsComponent),
  },
  {
    path: 'transfers',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/transfers/transfers.component').then((m) => m.TransfersComponent),
  },
  {
    path: 'goals',
    canActivate: [authGuard],
    loadComponent: () => import('./features/goals/goals.component').then((m) => m.GoalsComponent),
  },
  {
    path: 'incomes',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/incomes/incomes.component').then((m) => m.IncomesComponent),
  },
  {
    path: 'expenses',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/expenses/expenses.component').then((m) => m.ExpensesComponent),
  },
];
