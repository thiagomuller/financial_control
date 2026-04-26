import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Expense } from '../models/models';

@Injectable({ providedIn: 'root' })
export class ExpenseService {
  private http = inject(HttpClient);

  list(): Observable<Expense[]> {
    return this.http.get<Expense[]>('/api/expenses');
  }

  create(data: {
    name: string;
    value: number;
    bank_account_id: string;
    repeatable_day: number;
  }): Observable<Expense> {
    return this.http.post<Expense>('/api/expenses', data);
  }

  update(
    id: string,
    data: { name: string; value: number; bank_account_id: string; repeatable_day: number },
  ): Observable<Expense> {
    return this.http.put<Expense>(`/api/expenses/${id}`, data);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`/api/expenses/${id}`);
  }
}
