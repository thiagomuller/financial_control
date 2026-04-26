import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Goal } from '../models/models';

@Injectable({ providedIn: 'root' })
export class GoalService {
  private http = inject(HttpClient);

  list(): Observable<Goal[]> {
    return this.http.get<Goal[]>('/api/goals');
  }

  create(data: {
    name: string;
    start_date: string;
    end_date: string;
    interval_days: number;
    source_account_id: string;
    target_account_id: string;
    target_value: number;
  }): Observable<Goal> {
    return this.http.post<Goal>('/api/goals', data);
  }

  update(
    id: string,
    data: { name: string; end_date: string; interval_days: number; target_value: number },
  ): Observable<Goal> {
    return this.http.put<Goal>(`/api/goals/${id}`, data);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`/api/goals/${id}`);
  }
}
