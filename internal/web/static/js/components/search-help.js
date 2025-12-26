// Search Help Component for Snipo
export function initSearchHelp(Alpine) {
  Alpine.data('searchHelp', () => ({
    showHelp: false,
    
    toggleHelp() {
      this.showHelp = !this.showHelp;
    },
    
    closeHelp() {
      this.showHelp = false;
    }
  }));
}
