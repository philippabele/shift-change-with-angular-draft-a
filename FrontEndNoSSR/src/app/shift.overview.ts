export interface shift {
  datum: Date;
  uid: string;
  day: string;
  time: string;
  trade: boolean;
  search: search[];
}

export interface search {
  selected: boolean;
  name: string;
  offers: number;
}
