import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Income } from '../models/models';

@Injectable({ providedIn: 'root' })
export class IncomeService {
  private http = inject(HttpClient);

  list(): Observable<Income[]> { return this.http.get<Income[]>('/api/incomes'); }

  create(data: { name: string; value: number; bank_account_id: string; repeatable_day: number }): Observable<Income> {
    return this.http.post<Income>('/api/incomes', data);
  }

  update(id: string, data: { name: string; value: number; bank_account_id: string; repeatable_day: number }): Observable<Income> {
    return this.http.put<Income>(`/api/incomes/${id}`, data);
  }

  delete(id: string): Observable<void> { return this.http.delete<void>(`/api/incomes/${id}`); }
}
